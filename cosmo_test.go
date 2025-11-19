package cosmo

import (
	"fmt"
	"testing"
)

type DBService interface {
	Get() error
}

type SQLDBService struct {}

func (svc *SQLDBService) Get() error {
	fmt.Println("db service")
	return nil
}

type ToBind struct {
	DB DBService
}

func TestContainer(t *testing.T) {
	c := New()

	err := c.Add(func () DBService {
		return &SQLDBService{}
	})
	if err != nil {
		t.Error(err.Error())
	}

	err = c.Invoke(func (db DBService) {
		db.Get()
	})
	if err != nil {
		t.Error(err.Error())
	}

	var bnd ToBind
	if err = c.Bind(&bnd); err != nil {
		t.Error(err.Error())
	}

	bnd.DB.Get()
}
