package log

import (
	"errors"
	"io"
	"testing"

	api "github.com/lexnapoles/proglog/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newTestLog(t *testing.T) *Log {
	t.Helper()

	c := config()

	log, err := NewLog(t.TempDir(), c)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, log.Close())
	})

	return log
}

func config() Config {
	c := Config{}
	c.Segment.MaxStoreBytes = 32
	return c
}

func TestLogAppendRead(t *testing.T) {
	log := newTestLog(t)

	record := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := log.Append(record)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, record.Value, read.Value)
}

func TestLogOutOfRangeErr(t *testing.T) {
	log := newTestLog(t)

	read, err := log.Read(1)
	require.Nil(t, read)

	var apiErr api.ErrOffsetOutOfRange
	errors.As(err, &apiErr)

	require.Equal(t, uint64(1), apiErr.Offset)
}

func TestLogInitExisting(t *testing.T) {
	o, err := NewLog(t.TempDir(), config())
	require.NoError(t, err)

	record := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := o.Append(record)
		require.NoError(t, err)
	}

	require.NoError(t, o.Close())

	off, err := o.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = o.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

	n, err := NewLog(o.Dir, o.Config)
	require.NoError(t, err)
	defer n.Close()

	off, err = n.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = n.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)
}

func TestLogReader(t *testing.T) {
	log := newTestLog(t)

	record := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := log.Append(record)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	reader := log.Reader()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, record.Value, read.Value)
}

func TestLogTruncate(t *testing.T) {
	log := newTestLog(t)

	record := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := log.Append(record)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err)
}
