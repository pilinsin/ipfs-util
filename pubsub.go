package ipfs

import(
	"context"
	"time"

	iface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"

	"github.com/pilinsin/util"
)

type pubsub struct{
	api iface.CoreAPI
	ctx     context.Context
}
func (self *IPFS) PubSub() *pubsub{
	return &pubsub{self.api, self.ctx}
}
func (self *pubsub) Publish(data []byte, topic string) {
	self.api.PubSub().Publish(self.ctx, topic, data)
}
func (self *pubsub) Subscribe(topic string) iface.PubSubSubscription {
	sub, _ := self.api.PubSub().Subscribe(self.ctx, topic, options.PubSub.Discover(true))
	return sub
}
func (self *pubsub) Next(sub iface.PubSubSubscription) []byte {
	ctx, cancel := util.CancelTimerContext(5*time.Second)
	defer cancel()

	msg, err := sub.Next(ctx)
	if err != nil {
		return nil
	}
	return msg.Data()
}
func (self *pubsub) NextAll(sub iface.PubSubSubscription) [][]byte {
	var dataset [][]byte
	for {
		data := self.Next(sub)
		if data == nil {
			if len(dataset) > 0 {
				return dataset
			} else {
				return nil
			}
		}
		dataset = append(dataset, data)
	}
}
