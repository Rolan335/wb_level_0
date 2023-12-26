package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
)

var c *cache.Cache
var db *sql.DB

func main() {
	connStr := "postgresql://wb_user:Pa$$w0rd@localhost:5432/wb-database?sslmode=disable"
	db, _ = sql.Open("postgres", connStr)

	c = cache.New(7*24*time.Hour, 24*time.Hour)
	restoreCache()

	http.HandleFunc("/getOrderById", getOrderById)
	http.HandleFunc("/", PageHandler)

	natsURL := "nats://localhost:4222"
	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}

	var newOrder Order
	nc.Subscribe("order", func(m *nats.Msg) {
		err = json.Unmarshal(m.Data, &newOrder)
		if err != nil {
			fmt.Println(err)
		} else {
			id_order := writeToDatabase(newOrder)
			c.Set(strconv.Itoa(int(id_order)), newOrder, cache.DefaultExpiration)
		}
	})
	sender()

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}

func PageHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./web/index.html")
	t.Execute(w, nil)

	id := r.URL.Query().Get("id")

	res, err := http.Get("http://localhost:8082/getOrderById?id=" + id)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	resData, _ := io.ReadAll(res.Body)

	w.Write(resData)
}

func getOrderById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	order, ok := c.Get(id)
	if ok {
		json, _ := json.Marshal(order)
		w.Write(json)
	} else {
		w.Write([]byte("ID doesn't exists"))
	}

}

func restoreCache() {
	var id_order int64
	var delivery_id int64
	var payment_id int64

	//order
	rows, _ := db.Query(`SELECT * from "order" where date_created > current_date - interval '7' day`)
	for rows.Next() {
		var r Order
		_ = rows.Scan(&id_order, &delivery_id, &payment_id, &r.OrderUid, &r.TrackNumber, &r.Entry,
			&r.Locale, &r.InternalSignature, &r.CustomerId, &r.DeliveryService,
			&r.Shardkey, &r.SmId, &r.DateCreated, &r.OofShard)

		//delivery
		_ = db.QueryRow(`SELECT "name", phone, zip, city, address, region, email from delivery where id_delivery = $1`, delivery_id).Scan(
			&r.Delivery.Name, &r.Delivery.Phone, &r.Delivery.Zip, &r.Delivery.City, &r.Delivery.Address, &r.Delivery.Region, &r.Delivery.Email)

		//payment
		_ = db.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		from payment where id_payment = $1`, payment_id).Scan(
			&r.Payment.Transaction, &r.Payment.RequestId, &r.Payment.Currency, &r.Payment.Provider, &r.Payment.Amount, &r.Payment.PaymentDt,
			&r.Payment.Bank, &r.Payment.DeliveryCost, &r.Payment.GoodsTotal, &r.Payment.CustomFee)

		//items
		rowsItem, _ := db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status from items
		where order_id = $1`, id_order)
		for rowsItem.Next() {
			var i Item
			_ = rowsItem.Scan(&i.ChrtId, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale, &i.Size, &i.TotalPrice, &i.NmId, &i.Brand, &i.Status)
			r.Items = append(r.Items, i)
		}

		c.Set(strconv.Itoa(int(id_order)), r, cache.DefaultExpiration)
	}
}

func writeToDatabase(o Order) (id_order int32) {
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
