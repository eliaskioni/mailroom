package testsuite

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/nyaruka/gocommon/storage"
	"github.com/nyaruka/mailroom/config"
	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/runtime"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

/*var tableHashes = map[string]string{
	"channels_channel": "3587399bad341401f1880431c0bc772a",
	"contacts_contact": "0382ef6e58e260c0c76dcc84550e6793",
	"orgs_org":         "0f650bf7b9fb77ffa3ff0992be98da53",
	"tickets_ticketer": "6487a4aed61e16c3aa0d6cf117f58de3",
}*/

const MediaStorageDir = "_test_media_storage"
const SessionStorageDir = "_test_session_storage"

// Reset clears out both our database and redis DB
func Reset() {
	ResetDB()
	ResetRP()

	models.FlushCache()
	logrus.SetLevel(logrus.DebugLevel)
}

// Get returns the various runtime things a test might need
func Get() (context.Context, *runtime.Runtime, *sqlx.DB, *redis.Pool) {
	db := DB()
	rp := RP()
	rt := &runtime.Runtime{
		RP:             rp,
		DB:             db,
		ES:             nil,
		MediaStorage:   MediaStorage(),
		SessionStorage: SessionStorage(),
		Config:         config.NewMailroomConfig(),
	}

	/*for name, expected := range tableHashes {
		var actual string
	    must(db.Get(&actual, fmt.Sprintf(`SELECT md5(array_to_string(array_agg(t.* order by id), '|', '')) FROM %s t`, name)))
		if actual != expected {
			panic(fmt.Sprintf("table has mismatch for %s, expected: %s, got %s", name, expected, actual))
		}
	}*/

	return context.Background(), rt, db, rp
}

// ResetDB resets our database to our base state from our RapidPro dump
//
// mailroom_test.dump can be regenerated by running:
//   % python manage.py mailroom_db
//
// then copying the mailroom_test.dump file to your mailroom root directory
//   % cp mailroom_test.dump ../mailroom
func ResetDB() {
	db := sqlx.MustOpen("postgres", "postgres://mailroom_test:temba@localhost/mailroom_test?sslmode=disable&Timezone=UTC")
	defer db.Close()

	db.MustExec("drop owned by mailroom_test cascade")
	dir, _ := os.Getwd()

	// our working directory is set to the directory of the module being tested, we want to get just
	// the portion that points to the mailroom directory
	for !strings.HasSuffix(dir, "mailroom") && dir != "/" {
		dir = path.Dir(dir)
	}

	mustExec("pg_restore", "-h", "localhost", "-d", "mailroom_test", "-U", "mailroom_test", path.Join(dir, "./mailroom_test.dump"))
}

// DB returns an open test database pool
func DB() *sqlx.DB {
	return sqlx.MustOpen("postgres", "postgres://mailroom_test:temba@localhost/mailroom_test?sslmode=disable&Timezone=UTC")
}

// ResetRP resets our redis database
func ResetRP() {
	rc, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(fmt.Sprintf("error connecting to redis db: %s", err.Error()))
	}
	rc.Do("SELECT", 0)
	_, err = rc.Do("FLUSHDB")
	if err != nil {
		panic(fmt.Sprintf("error flushing redis db: %s", err.Error()))
	}
}

// RP returns a redis pool to our test database
func RP() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}
			_, err = conn.Do("SELECT", 0)
			return conn, err
		},
	}
}

// RC returns a redis connection, Close() should be called on it when done
func RC() redis.Conn {
	conn, err := redis.Dial("tcp", "localhost:6379")
	must(err)
	_, err = conn.Do("SELECT", 0)
	must(err)
	return conn
}

// MediaStorage returns our media storage for tests
func MediaStorage() storage.Storage {
	return storage.NewFS(MediaStorageDir)
}

// SessionStorage returns our session storage for tests
func SessionStorage() storage.Storage {
	return storage.NewFS(SessionStorageDir)
}

// ResetStorage clears our storage for tests
func ResetStorage() {
	must(os.RemoveAll(MediaStorageDir))
	must(os.RemoveAll(SessionStorageDir))
}

var resetDataSQL = `
DELETE FROM notifications_notification;
DELETE FROM request_logs_httplog;
DELETE FROM tickets_ticketevent;
DELETE FROM tickets_ticket;
DELETE FROM channels_channelcount;
DELETE FROM msgs_msg;
DELETE FROM campaigns_eventfire;
DELETE FROM contacts_contactimportbatch;
DELETE FROM contacts_contactimport;
DELETE FROM contacts_contacturn WHERE id >= 30000;
DELETE FROM contacts_contactgroup_contacts WHERE contact_id >= 30000;
DELETE FROM contacts_contact WHERE id >= 30000;
DELETE FROM contacts_contactgroupcount WHERE group_id >= 30000;
DELETE FROM contacts_contactgroup WHERE id >= 30000;

ALTER SEQUENCE tickets_ticket_id_seq RESTART WITH 1;
ALTER SEQUENCE msgs_msg_id_seq RESTART WITH 1;
ALTER SEQUENCE contacts_contact_id_seq RESTART WITH 30000;
ALTER SEQUENCE contacts_contacturn_id_seq RESTART WITH 30000;
ALTER SEQUENCE contacts_contactgroup_id_seq RESTART WITH 30000;`

// ResetData removes contact data not in the test database dump. Note that this function can't
// undo changes made to the contact data in the test database dump.
func ResetData(db *sqlx.DB) {
	db.MustExec(resetDataSQL)

	// because groups have changed
	models.FlushCache()
}

// utility function for running a command panicking if there is any error
func mustExec(command string, args ...string) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("error restoring database: %s: %s", err, string(output)))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
