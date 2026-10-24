package chapter06

type Product struct {
	Name string `gorm:"column:name;type:varchar(200);not null;comment:性名"`
	Price float64 `gorm:"column:price;type:decimal(10,2);default:0.00"`
	Stock int `gorm:"column:stock;type:int unsigned;default:0"`
	Status string `gorm:"column:status;type:varchar(200);index"`
	Sku string `gorm:"column:sku;type:varchar(200);uniqueIndex"`
}

func LearnCode() {

}