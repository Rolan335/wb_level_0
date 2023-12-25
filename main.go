package main

import (
	"database/sql"
	"encoding/json"

	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
)

var c *cache.Cache
var db *sql.DB

func main() {
	c = cache.New(7*24*time.Hour, 24*time.Hour)

	natsURL := "nats://localhost:4222"
	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	connStr := "postgresql://wb_user:Pa$$w0rd@localhost:5432/wb-database?sslmode=disable"
	db, _ = sql.Open("postgres", connStr)

	var newOrder Order
	nc.Subscribe("order", func(m *nats.Msg) {
		err = json.Unmarshal(m.Data, &newOrder)
		if err != nil {
			panic(err)
		}
		id_order := writeToDatabase(newOrder)
		c.Set(string(id_order), newOrder, cache.DefaultExpiration)
	})
	sender()
	restoreCache()
	for {

	}
}

func restoreCache() {

}

// val, found := c.Get(string(id_order))
// if found{
// fmt.Println(id_order,"\n",val)
// }

func writeToDatabase(o Order) (id_order int64) {

	defer db.Close()

	//delivery
	var id_delivery int64
	deliveryQuery := `INSERT INTO delivery (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id_delivery`
	_ = db.QueryRow(deliveryQuery, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email).Scan(&id_delivery)

	//payment
	var id_payment int64
	paymentQuery := `INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id_payment`
	_ = db.QueryRow(paymentQuery, o.Payment.Transaction, o.Payment.RequestId, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee).Scan(&id_payment)

	//order
	orderQuery := `INSERT INTO "order" (delivery_id, payment_id, order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) 
                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id_order`
	_ = db.QueryRow(orderQuery, id_delivery, id_payment, o.OrderUid, o.TrackNumber, o.Entry, o.Locale, o.CustomerId, o.DeliveryService, o.Shardkey, o.SmId, o.DateCreated, o.OofShard).Scan(&id_order)

	//items
	var id_items int64
	for _, item := range o.Items {
		itemQuery := `INSERT INTO items (order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) 
                      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id_items`
		_ = db.QueryRow(itemQuery, id_order, item.ChrtId, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status).Scan(&id_items)
	}

	return id_order
}
