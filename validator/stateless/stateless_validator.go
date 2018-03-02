package stateless

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/validation"
	"github.com/Ontology/eventbus/actor"
	vatypes "github.com/Ontology/validator/types"
)

type Validator interface {
	Register(poolId *actor.PID)
	UnRegister(poolId *actor.PID)
	VerifyType() vatypes.VerifyType
}

type validator struct {
	pid *actor.PID
	id  string
}

func NewValidator(id string) (Validator, error) {
	validator := &validator{id: id}
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, id)
	validator.pid = pid
	return validator, err
}

func (self *validator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Validator started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Validator stopping")
	case *actor.Restarting:
		log.Info("Validator Restarting")
	case *vatypes.CheckTx:
		log.Info("Validator receive tx")
		sender := context.Sender()
		errCode := validation.VerifyTransaction(&msg.Tx)

		response := &vatypes.StatelessCheckResponse{
			ErrCode: errCode,
			Hash:    msg.Tx.Hash(),
		}

		sender.Tell(response)
	case *vatypes.UnRegisterAck:
		context.Self().Stop()
	default:
		log.Info("hahaha", context.Sender())
		log.Info("Unknown msg type", msg)
	}

}

func (self *validator) VerifyType() vatypes.VerifyType {
	return vatypes.Stateless
}

func (self *validator) Register(poolId *actor.PID) {
	poolId.Tell(&vatypes.RegisterValidator{
		Sender: self.pid,
		Type:   self.VerifyType(),
		Id:     self.id,
	})
}

func (self *validator) UnRegister(poolId *actor.PID) {
	poolId.Tell(&vatypes.UnRegisterValidator{
		Id: self.id,
	})

}