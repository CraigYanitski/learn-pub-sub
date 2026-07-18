package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		res := gs.HandleMove(move)
		defer func() {
			if res == gamelogic.MoveOutcomeMakeWar {
				return
			}
			fmt.Print("> ")
		} ()
		fmt.Println("result: ", res)
		switch res {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			//time.Sleep(1*time.Second)
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(recognition gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		res, winner, loser := gs.HandleWar(recognition)
		//time.Sleep(500*time.Millisecond)
		var msg string
		switch res {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				gs.GetUsername(),
				msg,
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				gs.GetUsername(),
				msg,
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := publishGameLog(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.GameLogSlug, gs.GetUsername()),
				gs.GetUsername(),
				msg,
			)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Unknown war outcome, discarding")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(ch *amqp.Channel, exchange, key, user, log string) error {
	gl := routing.GameLog{
		CurrentTime: time.Now(),
		Message: log,
		Username: user,
	}
	err := pubsub.PublishGob(ch, exchange, key, gl)
	if err != nil {
		return err
	}
	return nil
}

