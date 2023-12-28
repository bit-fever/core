//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package msg

import (
	"context"
	"encoding/json"
	"github.com/bit-fever/core"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"time"
)

//=============================================================================

const ExInventoryUpdates = "bf.inventory.updates"
const QuInventoryUpdatesToPortfolio = "bf.inventory.updates:portfolio"

var channel *amqp.Channel

//=============================================================================

func InitMessaging(cfg *core.Messaging) {

	slog.Info("Starting messaging...")
	url := "amqp://"+ cfg.Username + ":" + cfg.Password + "@" + cfg.Address + "/"

	conn, err := amqp.Dial(url)
	if err != nil {
		core.ExitWithMessage("Failed to connect to the messaging system: "+ err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		core.ExitWithMessage("Failed to get a channel from the messaging system: "+ err.Error())
	}

	channel = ch

	createExchange(ExInventoryUpdates)
	createQueue   (QuInventoryUpdatesToPortfolio)
	bindQueue     (ExInventoryUpdates, QuInventoryUpdatesToPortfolio)
}

//=============================================================================

func PublishToExchange(exchange string, message any) error {
	body, err := json.Marshal(&message)
	if err != nil {
		slog.Error("Error marshalling message", "error", err.Error())
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(ctx, exchange, "", false, false,
		amqp.Publishing{
			ContentType: "text/json",
			Body:        body,
		})

	if err != nil {
		core.ExitWithMessage("Cannot publish to exchange '"+ exchange +"' : "+ err.Error())
	}

	return err
}

//=============================================================================

func SendMessage(exchange string, origin int, msgType int, source string, entity any) error {
	body, err := json.Marshal(entity)
	if err != nil {
		slog.Error("Error marshalling message", "error", err.Error())
		return err
	}

	message := &Message{
		Origin: origin,
		Type  : msgType,
		Source: source,
		Entity: body,
	}

	return PublishToExchange(exchange, message)
}

//=============================================================================

func ReceiveMessages(queue string, handler func(m *Message) bool) {
	messages, err := channel.Consume(queue,"",false,false,false,false,nil)

	if err != nil {
		core.ExitWithMessage("Cannot create the consumer channel for '"+ queue +"' : "+ err.Error())
	}

	for d := range messages {
		msg := Message{}
		err := json.Unmarshal(d.Body, &msg)

		if err != nil {
			slog.Error("Error unmarshalling message. Rejecting.", "error", err.Error())
			err = d.Reject(false)
			if err != nil {
				slog.Error("Cannot reject message!", "error", err.Error())
			}
			continue
		}

		if handler(&msg) {
			err = d.Ack(false)
		} else {
			err = d.Nack(false, true)
		}

		if err != nil {
			slog.Error("Cannot [N]acknowledge message!", "error", err.Error())
		}
	}
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func createExchange(name string) {
	err := channel.ExchangeDeclare(name,"fanout",true,false,false,false,nil)

	if err != nil {
		core.ExitWithMessage("Cannot create the '"+ name +"' exchange in the messaging system: "+ err.Error())
	}
}

//=============================================================================

func createQueue(name string) {
	_, err := channel.QueueDeclare(name,true,false,false,false,nil)

	if err != nil {
		core.ExitWithMessage("Cannot create the '"+ name +"' queue in the messaging system: "+ err.Error())
	}
}

//=============================================================================

func bindQueue(exchange, queue string) {
	err := channel.QueueBind(queue,"",exchange,false,nil)

	if err != nil {
		core.ExitWithMessage("Cannot bind queue '"+ queue +"' to the exchange: "+ err.Error())
	}
}

//=============================================================================
