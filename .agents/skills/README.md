# Подключенные скиллы

В данном проекте используются следующие скиллы для разработки на Go:

## 1. golang-patterns
**Путь:** `/home/user/.agents/skills/golang-patterns/SKILL.md`

### Ключевые принципы:
- Простота и ясность кода
- Использование zero value
- Принцип: "Accept interfaces, return structs"
- Паттерны обработки ошибок с wrapping
- Паттерны конкурентности: worker pool, errgroup, graceful shutdown
- Структура проекта: cmd/, internal/, pkg/
- Functional Options pattern
- Избегание goroutine leaks

### Команды:
```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
```

## 2. software-architecture
**Путь:** `/home/user/.agents/skills/software-architecture/SKILL.md`

### Ключевые принципы:
- Clean Architecture & DDD
- Early return pattern
- Library-first подход
- Избегание NIH синдрома
- Separation of Concerns
- Декомпозиция компонентов < 200 строк
- Доменные имена вместо generic (utils, helpers, common)

### Архитектура слоев:
- Handler → Service → Repository
- Domain entities отделены от infrastructure

## 3. effective-go
**Путь:** `/home/user/.agents/skills/effective-go/SKILL.md`

### Ключевые правила:
- **Форматирование**: всегда gofmt
- **Именование**: MixedCaps для экспортируемых, mixedCaps для внутренних
- **Обработка ошибок**: всегда проверять, возвращать, не паниковать
- **Конкурентность**: делиться памятью через коммуникацию (каналы)
- **Интерфейсы**: small (1-3 метода), accept interfaces, return structs
- **Документация**: документировать все экспортируемые символы

## Применение в проекте

При реализации AI Assistant используем:
1. Паттерн слоистой архитектуры из software-architecture
2. Go-идиомы из golang-patterns и effective-go
3. Структуру: `cmd/server`, `internal/{handler,service,repository,model,config}`
4. Minimal interfaces для зависимостей
5. Proper error wrapping с контекстом
6. Context для cancellation и graceful shutdown

---

*Скиллы загружены: 2024-01-15*
