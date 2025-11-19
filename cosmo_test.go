package cosmo

import (
	"fmt"
	"testing"
)

type DBService interface {
	Get() error
}

type Config struct {
	URL string
}

type SQLDBService struct{
	Config Config
}

func (svc *SQLDBService) Get() error {
	fmt.Printf("Database URL: %s\n", svc.Config.URL)
	return nil
}

type ToBind struct {
	DB DBService
}

func TestContainer(t *testing.T) {
	c := New()

	singletonConstructorCallTimes := 0

	err := c.AddSingleton(func() Config {
		singletonConstructorCallTimes++
		return Config{
			URL: "sqlite://test.db",
		}
	})
	if err != nil {
		t.Error(err.Error())
	}

	err = c.Add(func(cfg Config) DBService {
		return &SQLDBService{
			Config: cfg,
		}
	})
	if err != nil {
		t.Error(err.Error())
	}

	err = c.Invoke(func(db DBService) {
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

	if singletonConstructorCallTimes != 1 {
		t.Errorf("Singleton constructor was called %d times", singletonConstructorCallTimes)
	}
}
