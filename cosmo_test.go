package cosmo

import (
	"testing"
)

const DBURL string = "sqlite://test.db"

type DBService interface {
	Get() error
}

type Config struct {
	URL string
}

type SQLDBService struct {
	Config Config
}

func (svc *SQLDBService) Get() error {
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
			URL: DBURL,
		}
	})
	if err != nil {
		t.Error(err.Error())
	}

	err = c.Add(func(cfg Config) DBService {
		if cfg.URL != DBURL {
			t.Errorf("wrong value injected into Config")
		}
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

	if err = bnd.DB.Get(); err != nil {
		t.Error(err.Error())
	}
	if singletonConstructorCallTimes != 1 {
		t.Errorf("Singleton constructor was called %d times", singletonConstructorCallTimes)
	}
}

func TestConfigure(t *testing.T) {
	c := New()
	c.Configure("DBConfig", func() Config {
		return Config{
			URL: DBURL,
		}
	})
	cfg, ok := c.Get("DBConfig").(Config)
	if !ok {
		t.Errorf("could not cast key DBConfig to Config")
	}
	if cfg.URL != DBURL {
		t.Errorf("invalid injected value on Config")
	}
	c.Configure("DBService", func(cfg Config) DBService {
		return &SQLDBService{
			Config: cfg,
		}
	})
	service, ok := c.Get("DBService").(DBService)
	if !ok {
		t.Errorf("could not cast key DBService to DBService")
	}
	service.Get()
}

func TestContext(t *testing.T) {
	c := New()
	c.Configure("DBConfig", func() Config {
		return Config{
			URL: DBURL,
		}
	})
	ctx := c.Context()
	cfg, ok := Context(ctx, "DBConfig").(Config)
	if !ok {
		t.Error("could not get DBConfig from CosmoContext")
	}
	if cfg.URL != DBURL {
		t.Error("wrong data inject in Config")
	}
}

func TestResolveNotAddedService(t *testing.T) {
	c := New()
	c.Configure("DBConfig", func() Config {
		return Config{
			URL: DBURL,
		}
	})
	cfg := c.Get("Config")
	if cfg != nil {
		t.Error("not nil return from not added service")
	}
}

func TestNotFuncConstructor(t *testing.T) {
	c := New()
	if err := c.Configure("DBConfig", Config{}); err == nil {
		t.Error("invalid validation for ctor on Configure")
	}

	if err := c.Add(Config{}); err == nil {
		t.Error("invalid validation for ctor on Add")
	}

	if err := c.AddSingleton(Config{}); err == nil {
		t.Error("invalid validation for ctor on AddWithSingleton")
	}
}

func ExampleContainer() {
	c := New()
	c.Configure("DBConfig", func() Config {
		return Config{
			URL: DBURL,
		}
	})
	c.Configure("DBService", func(cfg Config) DBService {
		return &SQLDBService{
			Config: cfg,
		}
	})
	service, ok := c.Get("DBService").(DBService)
	if !ok {
		panic("Dependency injection error")
	}
	service.Get()
}
