## 1. Базовый принцип (очень важно)

> **Экспортируй поведение, скрывай данные**

В Go продакшн-код ломается не из-за DDD, а из-за:

* экспорта структур
* передачи указателей наружу
* прямого доступа к состоянию

### ❌ Плохо

```go
type User struct {
    ID    int64
    Email string
    Role  string
}
```

Любой пакет может:

* мутировать
* зависеть от структуры
* сломаться при изменении

### ✅ Правильно

```go
type User struct {
    id    int64
    email string
    role  string
}

func (u User) ID() int64     { return u.id }
func (u User) Email() string { return u.email }
```

**Инварианты живут внутри пакета.**

---

## 3. Interface{} — забудь

Скажу жёстко:
**`interface{}` в продакшне — почти всегда архитектурная ошибка.**

### Почему:

* теряется контракт
* нет compile-time гарантий
* код становится динамическим мусором

### Единственный допустимый кейс

* инфраструктура (json, sql, logging)
* generic-обвязки

В бизнес-коде:

> **никогда**

---

## 4. Interfaces: где и как объявлять

### Золотое правило Go

> **Интерфейс объявляет тот, кто его использует**

Не тот, кто реализует.

### ❌ Антипаттерн

```go
// user/repository.go
type Repository interface {
    Save(ctx context.Context, u *User) error
}
```

### ✅ Правильно

```go
// app/user/service.go
type userRepository interface {
    Save(ctx context.Context, u *user.User) error
}
```

Почему:

* интерфейс минимален
* реализация свободна
* легче менять хранилище

---

## 5. Указатели vs значения

### Структуры домена

* **immutable-подход**
* наружу — значения
* внутри пакета — указатели

```go
func NewUser(id int64, email string) (*User, error) {
    // проверки
    return &User{id: id, email: email}, nil
}
```

Снаружи:

```go
func (u User) Email() string
```

### DTO / Transport

* можно указатели
* можно экспортные поля
* они тупые

---

## 6. Границы модулей (пакетов)

Продакшн-границы строятся **не по слоям**, а по **фичам / сабдоменам**.

Не:

```
/handlers
/services
/repositories
```

А:

```
/user
/billing
/auth
```

---

## 7. Реальная структура папок (production-ready)

Пример для backend API:

```
/cmd
  /api
    main.go

/internal
  /app
    /user
      service.go
      usecase.go
      ports.go
    /auth
      service.go

  /domain
    /user
      user.go
      errors.go
      rules.go

  /infra
    /http
      /user
        handler.go
    /postgres
      /user
        repository.go

  /pkg
    /clock
    /logger

/go.mod
```

---

## 8. Что где живёт

### `/domain`

* бизнес-сущности
* инварианты
* **никаких интерфейсов инфраструктуры**
* чистый Go

### `/app`

* use cases
* orchestration
* интерфейсы (ports)
* транзакционные границы

### `/infra`

* БД
* HTTP
* Kafka
* Redis
* **зависит от app, но не наоборот**

### `/cmd`

* точка сборки
* wiring зависимостей

---

## 9. Связи (направление зависимостей)

```
cmd
 ↓
infra → app → domain
          ↑
        interfaces
```

**domain не знает ни о чём**