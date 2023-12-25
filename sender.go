package main

import(
	"github.com/nats-io/nats.go"
)

func sender() {
	natsURL := "nats://localhost:4222"
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return
	}
	defer nc.Close()
	nc.Publish("order",[]byte(`{"order_uid":"b563feb7b2b84b6test","track_number":"WBILMTESTTRACK","entry":"WBIL","delivery":{"name":"Test Testov 2 items","phone":"+9720000000","zip":"2639809","city":"Kiryat Mozkin","address":"Ploshad Mira 15","region":"Kraiot","email":"test@gmail.com"},"payment":{"transaction":"b563feb7b2b84b6test","request_id":"","currency":"USD","provider":"wbpay","amount":1817,"payment_dt":1637907727,"bank":"alpha","delivery_cost":1500,"goods_total":317,"custom_fee":0},"items":[{"chrt_id":9934930,"track_number":"ABILMTESTTRACK","price":453,"rid":"ab4219087a764ae0btest","name":"Mascaras","sale":30,"size":"0","total_price":317,"nm_id":2389212,"brand":"Vivienne Sabo 2","status":202},  {
      "chrt_id": 9934930,
      "track_number": "ABILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo 2",
      "status": 202
    }],"locale":"en","internal_signature":"","customer_id":"test","delivery_service":"meest","shardkey":"9","sm_id":99,"date_created":"2023-12-20T06:22:19Z","oof_shard":"1"}`))
}