package kafka

import (
	"L0WB/internal/models"
	"fmt"
	"strings"
)

func ValidDelivery(d *models.Delivery) error {
	if d.Name == "" {
		return fmt.Errorf("invalid Name in delivery: %s", d.Name)
	}
	if !strings.HasPrefix(d.Phone, "+") {
		return fmt.Errorf("invalid Phone in delivery: %s", d.Phone)
	}
	if len(d.Zip) <= 1 {
		return fmt.Errorf("invalid Zip in delivery: %s", d.Zip)
	}
	if len(d.City) <= 1 {
		return fmt.Errorf("invalid City in delivery: %s", d.Zip)
	}
	if len(d.Address) <= 1 {
		return fmt.Errorf("invalid Address in delivery: %s", d.Zip)
	}
	if len(d.Region) <= 1 {
		return fmt.Errorf("invalid Region in delivery: %s", d.Zip)
	}
	if !strings.Contains(d.Email, "@") {
		return fmt.Errorf("invalid Email in delivery: %s", d.Email)
	}
	return nil
}

func ValidPayment(p *models.Payment) error {
	if len(p.TransactionNumber) == 0 {
		return fmt.Errorf("invalid TransactionNumber in payment: %s", p.TransactionNumber)
	}
	if len(p.Currency) != 3 {
		return fmt.Errorf("invalid Currency in payment: %s", p.Currency)
	}
	if p.Provider == "" {
		return fmt.Errorf("invalid Provider in payment: %s", p.Provider)
	}
	if p.Amount <= 0 {
		return fmt.Errorf("invalid Amount in payment: %s", p.Amount)
	}
	if p.PaymentDT <= 0 {
		return fmt.Errorf("invalid PaymentDT in payment: %s", p.PaymentDT)
	}
	if p.Bank == "" {
		return fmt.Errorf("invalid Bank in payment: %s", p.Bank)
	}
	if p.DeliveryCost <= 0 {
		return fmt.Errorf("invalid DeliveryCost in payment: %s", p.DeliveryCost)
	}
	if p.GoodsTotal <= 0 {
		return fmt.Errorf("invalid GoodsTotal in payment: %s", p.GoodsTotal)
	}
	if p.CustomFee <= 0 {
		return fmt.Errorf("invalid CustomFee in payment: %s", p.CustomFee)
	}
	return nil
}

func ValidItem(i *models.Item) error {
	if i.ChrtID <= 0 {
		return fmt.Errorf("invalid ChrtID in item: %d", i.ChrtID)
	}
	if len(i.TrackNumber) != 13 {
		return fmt.Errorf("invalid TrackNumber in item: %s (expected 13 characters)", i.TrackNumber)
	}
	if i.Price <= 0 {
		return fmt.Errorf("invalid Price in item: %d", i.Price)
	}
	if len(i.RID) != 19 {
		return fmt.Errorf("invalid RID in item: %s (expected 19 characters)", i.RID)
	}
	if i.ItemName == "" {
		return fmt.Errorf("invalid ItemName in item: %s", i.ItemName)
	}
	if i.Sale < 0 || i.Sale > 100 {
		return fmt.Errorf("invalid Sale in item: %d (expected 0-100)", i.Sale)
	}
	if i.ItemSize == "" {
		return fmt.Errorf("invalid ItemSize in item: %s (expected 2 characters)", i.ItemSize)
	}
	if i.TotalPrice <= 0 {
		return fmt.Errorf("invalid TotalPrice in item: %d", i.TotalPrice)
	}
	if i.NmID <= 0 {
		return fmt.Errorf("invalid NmID in item: %d", i.NmID)
	}
	if i.Brand == "" {
		return fmt.Errorf("invalid Brand in item: %s", i.Brand)
	}
	validStatuses := map[int]bool{200: true, 400: true, 404: true}
	if !validStatuses[i.Status] {
		return fmt.Errorf("invalid Status in item: %d (expected 200, 400 or 404)", i.Status)
	}
	return nil
}

func ValidData(order *models.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("invalid OrderUID in order: %s", order.OrderUID)
	}
	if order.TrackNumber == "" {
		return fmt.Errorf("invalid TrackNumber in order: %s", order.TrackNumber)
	}
	if order.Entry == "" {
		return fmt.Errorf("invalid Entry in order: %s", order.Entry)
	}
	if err := ValidDelivery(&order.Delivery); err != nil {
		return err
	}
	if err := ValidPayment(&order.Payment); err != nil {
		return err
	}
	if len(order.Locale) != 2 {
		return fmt.Errorf("invalid Locale: %s", order.Locale)
	}
	if order.CustomerID == "" {
		return fmt.Errorf("invalid CustomerID: %s", order.CustomerID)
	}
	if order.DeliveryService == "" {
		return fmt.Errorf("invalid DeliveryService: %s", order.DeliveryService)
	}
	if order.Shardkey == "" {
		return fmt.Errorf("invalid Shardkey: %s", order.Shardkey)
	}
	if order.SmID < 0 {
		return fmt.Errorf("invalid SmID: %d", order.SmID)
	}
	if order.OofShard == "" {
		return fmt.Errorf("invalid OofShard: %s", order.OofShard)
	}
	for _, i := range order.Items {
		if err := ValidItem(&i); err != nil {
			return err
		}
	}
	return nil
}
