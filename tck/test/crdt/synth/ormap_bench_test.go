package synth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudstateio/go-support/cloudstate/encoding"
	"github.com/cloudstateio/go-support/cloudstate/entity"
	"github.com/cloudstateio/go-support/tck/crdt"
)

func BenchmarkCRDTORMap(b *testing.B) {
	s := newServer(&testing.T{})
	defer s.teardown()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	entityID := "ormap-1"
	command := "ProcessORMap"
	p := newProxy(ctx, s)
	defer p.teardown()
	p.init(&entity.CrdtInit{ServiceName: serviceName, EntityId: entityID})

	tr := tester{s.t}
	var sum uint64
	b.Run("ORMap", func(b *testing.B) {
		b.ReportAllocs()
		defer func() {
			fmt.Println(sum)
		}()
		inc0 := uint64(1)
		for i := 0; i < b.N; i++ {
			switch m := p.command(entityID, command,
				ormapRequest(&crdt.ORMapActionRequest{
					EntryKey: encoding.String("niner"),
					Request: &crdt.ORMapActionRequest_GCounterRequest{
						GCounterRequest: gcounterRequest(&crdt.GCounterIncrement{Value: inc0}),
					},
				}),
			).Message.(type) {
			case *entity.CrdtStreamOut_Reply:
				sum += inc0
				var r crdt.ORMapResponse
				tr.toProto(m.Reply.GetClientAction().GetReply().GetPayload(), &r)
				var state entity.CrdtState_Gcounter
				err := encoding.DecodeStruct(r.GetEntries().Values[0].GetValue(), &state)
				if err != nil {
					tr.t.Fatal(err)
				}
			default:
				tr.unexpected(m)
			}
		}
	})
}
