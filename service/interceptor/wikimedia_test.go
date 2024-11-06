package interceptor

import (
	"context"
	"github.com/awakari/source-sse/service/writer"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

type writerMock struct {
	chEvt chan<- *pb.CloudEvent
}

func newWriterMock(chEvt chan<- *pb.CloudEvent) writer.Service {
	return writerMock{
		chEvt: chEvt,
	}
}

func (w writerMock) Close() error {
	return nil
}

func (w writerMock) Write(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	w.chEvt <- evt
	return
}

func TestWikiMedia_Handle(t *testing.T) {
	chEvt := make(chan *pb.CloudEvent, 1)
	defer close(chEvt)
	w := newWriterMock(chEvt)
	w = writer.NewLogging(w, slog.Default())
	i := NewWikiMedia(w, "default", "com_awakari_sse_v1")
	cases := map[string]struct {
		in      string
		matches bool
		err     error
	}{
		"ok": {
			matches: true,
			in: `{
  "$schema": "/mediawiki/recentchange/1.0.0",
  "meta": {
    "uri": "https://commons.wikimedia.org/wiki/File:LL-Q150_(fra)-Lepticed7-d%C3%A9cro%C3%AEtre.wav",
    "request_id": "e146cffa-8ce0-45c2-b94d-e2f787066200",
    "id": "56b55034-a8db-4c5c-8f20-64b022bfedb9",
    "dt": "2024-11-06T08:56:22Z",
    "domain": "commons.wikimedia.org",
    "stream": "mediawiki.recentchange",
    "topic": "codfw.mediawiki.recentchange",
    "partition": 0,
    "offset": 1228489531
  },
  "id": 2648099400,
  "type": "edit",
  "namespace": 6,
  "title": "File:LL-Q150 (fra)-Lepticed7-décroître.wav",
  "title_url": "https://commons.wikimedia.org/wiki/File:LL-Q150_(fra)-Lepticed7-d%C3%A9cro%C3%AEtre.wav",
  "comment": "/* wbeditentity-update:0| */ automatically adding [[Commons:Structured data|structured data]] based on file information",
  "timestamp": 1730883382,
  "user": "SchlurcherBot",
  "bot": true,
  "notify_url": "https://commons.wikimedia.org/w/index.php?diff=953382381&oldid=754324688&rcid=2648099400",
  "minor": false,
  "patrolled": true,
  "length": {
    "old": 4940,
    "new": 6550
  },
  "revision": {
    "old": 754324688,
    "new": 953382381
  },
  "server_url": "https://commons.wikimedia.org",
  "server_name": "commons.wikimedia.org",
  "server_script_path": "/w",
  "wiki": "commonswiki",
  "parsedcomment": "‎<span dir=\"auto\"><span class=\"autocomment\">Changed an entity: </span></span> automatically adding <a href=\"/wiki/Commons:Structured_data\" title=\"Commons:Structured data\">structured data</a> based on file information"
}`,
		},
		"mismatch": {
			in: `{
  "$schema": "invalid",
  "meta": {
    "uri": "https://commons.wikimedia.org/wiki/File:LL-Q150_(fra)-Lepticed7-d%C3%A9cro%C3%AEtre.wav",
    "request_id": "e146cffa-8ce0-45c2-b94d-e2f787066200",
    "id": "56b55034-a8db-4c5c-8f20-64b022bfedb9",
    "dt": "2024-11-06T08:56:22Z",
    "domain": "commons.wikimedia.org",
    "stream": "mediawiki.recentchange",
    "topic": "codfw.mediawiki.recentchange",
    "partition": 0,
    "offset": 1228489531
  },
  "id": 2648099400,
  "type": "edit"
}`,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var raw map[string]any
			err := sonic.UnmarshalString(c.in, &raw)
			require.Nil(t, err)
			matches, err := i.Handle(context.TODO(), "src0", nil, raw)
			assert.Equal(t, c.matches, matches)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil && c.matches {
				evt := <-chEvt
				assert.NotNil(t, evt)
			}
		})
	}
}
