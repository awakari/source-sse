package interceptor

import (
	"context"
	"fmt"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service/writer"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/r3labs/sse/v2"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
)

type wikiMedia struct {
	w       writer.Service
	groupId string
	et      string
}

const keyWikiMediaSchema = "$schema"
const keyWikiMediaLength = "length"
const keyWikiMediaLogActionComment = "log_action_comment"
const keyWikiMediaNew = "new"
const keyWikiMediaNotifyUrl = "notify_url" // -> objecturl
const keyWikiMediaParsedComment = "parsedcomment"
const keyWikiMediaRevision = "revision"
const keyWikiMediaServerUrl = "server_url"
const keyWikiMediaTimestamp = "timestamp" // e.g. 1730883383
const keyWikiMediaTitle = "title"
const keyWikiMediaTitleUrl = "title_url"
const keyWikiMediaType = "type" // -> action
const keyWikiMediaUser = "user" // -> subject

const valWikiMediaSchema = "/mediawiki/recentchange/1.0.0"

func NewWikiMedia(w writer.Service, groupId, et string) Interceptor {
	return wikiMedia{
		w:       w,
		groupId: groupId,
		et:      et,
	}
}

func (wm wikiMedia) Handle(ctx context.Context, src string, ssEvt *sse.Event, raw map[string]any) (matches bool, err error) {

	schema, schemaOk := raw[keyWikiMediaSchema]
	matches = schemaOk && schema == valWikiMediaSchema
	if matches {

		txtRaw, txtOk := raw[keyWikiMediaParsedComment]
		if !txtOk || txtRaw == "" {
			txtRaw, txtOk = raw[keyWikiMediaLogActionComment]
		}
		var txt string
		if txtOk {
			txt, txtOk = txtRaw.(string)
		}
		if txtOk {

			evt := &pb.CloudEvent{
				Id:          ksuid.New().String(),
				Source:      src,
				SpecVersion: model.CeSpecVersion,
				Type:        wm.et,
				Attributes: map[string]*pb.CloudEventAttributeValue{
					model.CeKeySchema: {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: valWikiMediaSchema,
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: txt,
				},
			}

			ts, tsOk := raw[keyWikiMediaTimestamp]
			var tsUnixSeconds int64
			if tsOk {
				tsOk = true // assume
				switch tst := ts.(type) {
				case int64:
					tsUnixSeconds = tst
				case int32:
					tsUnixSeconds = int64(tst)
				case float32:
					tsUnixSeconds = int64(tst)
				case float64:
					tsUnixSeconds = int64(tst)
				default:
					tsOk = false
				}
			}
			if tsOk {
				evt.Attributes[model.CeKeyTime] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeTimestamp{
						CeTimestamp: &timestamppb.Timestamp{
							Seconds: tsUnixSeconds,
						},
					},
				}
			}

			objUrl, objUrlOk := raw[keyWikiMediaNotifyUrl]
			if !objUrlOk {
				objUrl, objUrlOk = raw[keyWikiMediaTitleUrl]
			}
			if objUrlOk {
				evt.Attributes[model.CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeUri{
						CeUri: objUrl.(string),
					},
				}
			}

			title, titleOk := raw[keyWikiMediaTitle]
			if titleOk {
				title, titleOk = title.(string)
			}
			if titleOk {
				evt.Attributes[model.CeKeyTitle] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: title.(string),
					},
				}
			}

			typ, typOk := raw[keyWikiMediaType]
			if typOk {
				typ, typOk = typ.(string)
			}
			if typOk {
				// rename new to create
				if typ == "new" {
					typ = "create"
				}
				evt.Attributes[model.CeKeyAction] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: typ.(string),
					},
				}
			}

			subj, subjOk := raw[keyWikiMediaUser]
			if subjOk {
				subj, subjOk = subj.(string)
			}
			if subjOk {
				evt.Attributes[model.CeKeySubject] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: subj.(string),
					},
				}
			}

			length, lengthOk := raw[keyWikiMediaLength]
			if lengthOk {
				length, lengthOk = length.(map[string]any)
			}
			if lengthOk {
				lengthNew, lengthNewOk := length.(map[string]any)[keyWikiMediaNew]
				if lengthNewOk {
					lengthNewOk = true // assume
					switch tln := lengthNew.(type) {
					case int:
						lengthNew = int32(tln)
					case int32:
						lengthNew = tln
					case int64:
						lengthNew = int32(tln)
					case float32:
						lengthNew = int32(tln)
					case float64:
						lengthNew = int32(tln)
					default:
						lengthNewOk = false
					}
				}
				if lengthNewOk {
					evt.Attributes[model.CeKeyLength] = &pb.CloudEventAttributeValue{
						Attr: &pb.CloudEventAttributeValue_CeInteger{
							CeInteger: lengthNew.(int32),
						},
					}
				}
			}

			rev, revOk := raw[keyWikiMediaRevision]
			if revOk {
				rev, revOk = rev.(map[string]any)
			}
			if revOk {
				revNew, revNewOk := rev.(map[string]any)[keyWikiMediaNew]
				if revNewOk {
					switch tr := revNew.(type) {
					case int:
						revNew = strconv.Itoa(tr)
					case int32:
						revNew = strconv.Itoa(int(tr))
					case int64:
						revNew = strconv.Itoa(int(tr))
					case float32:
						revNew = strconv.Itoa(int(tr))
					case float64:
						revNew = strconv.Itoa(int(tr))
					}
					evt.Attributes[model.CeKeyRevision] = &pb.CloudEventAttributeValue{
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: revNew.(string),
						},
					}
				}
			}

			var userId string
			serverUrl, serverUrlOk := raw[keyWikiMediaServerUrl]
			if serverUrlOk {
				userId, serverUrlOk = serverUrl.(string)
			}
			if !serverUrlOk {
				userId = src
			}

			if evt.GetTextData() == "" {
				err = fmt.Errorf("empty event text content, source: %s, data: %s", src, string(ssEvt.Data))
			}
			if objUrl, objUrlOk := evt.Attributes[model.CeKeyObjectUrl]; objUrl.GetCeUri() == "" || !objUrlOk {
				err = fmt.Errorf("empty event object url, source: %s, data: %s", src, string(ssEvt.Data))
			}
			if err == nil {
				err = wm.w.Write(ctx, evt, wm.groupId, userId)
			}
		}
	}
	return
}
