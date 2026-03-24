<h1> Система для работы с постами и комментариями </h1>

В этом проекте реализована система для работы с постами и комментариями аналогичная блогам/формам

Вдохновлено примером из библиотеки https://github.com/99designs/gqlgen/tree/master/_examples/mini-habr-with-subscriptions

## Реализованно:
- Хранение данных в PostgreSQL с использованием sqlc для генерации типизированного кода;
- Генерация GraphQL‑схемы и резолверов с помощью gqlgen;
- Поддержка in‑memory хранилища для быстрого прототипирования и тестов;
- Древовидные комментарии с помощью path (массив идентификаторов) для эффективной сортировки и выборки;
- Подписка на новые комментарии через GraphQL Subscriptions (WebSocket);
- Docker Compose для быстрого развертывания

## Что может быть улучшено:
- Тестирование (сейчас только unit в memdb и интеграционные с использованием membd в сервисе)
- Более надёжные подписки (сейчас подписки хранятся в памяти, они потеряются при перезапуске/падении приложения). Можно хранить во внешнем хранилище; если комментарии приходят быстро - канал может переполнится и начать отбрасывать самые новые комментарии, что не очень хорошо.
- Работа над алгоритмом выборки из базы - сейчас строится дерево по выбранным корневым комментариям - спасает нас от N+1, но если у корневых комментариев будет очень много ответов можем засыпаться. Рассматривал вариант с dataLoader, но так как в задании прописано, что может быть большая глубина, то для комментариев с большой вложенности придётся делать очень много запросов в отличие от варианта с деревом. В реальных системах обычно есть какое-то адекватное ограничение на вложенность, поэтому там dataLoader будет лучше. В нашем же случае момент спорный. 
- Кастомные ошибки(сейчас просто fmt.Errorf()
- Настроить логирование
- Метрики, трейсинг

Стек технологий

    Go 1.26.1

    GraphQL – gqlgen

    PostgreSQL 17

    pgx – драйвер для работы с PostgreSQL

    sqlc – генерация кода для работы с БД

    goose – миграции

    Docker / Docker Compose

    testify – для тестирования

Как запустить у себя:
1. Клонируйте репозиторий
   ```
   git clone https://github.com/ganzhinza/postsys.git
   ```
2. Настроить переменные окружения(через файл .env)
3. Запуск через Docker Compose
   ```
   docker-compose up
   ```
   После запуска GraphQL Playground будет доступен по адресу: http://localhost:8080/
   P.S. Для продакшена лучше отключать, но удобно для тестирования

Примеры запросов:
- Получить все посты
  ```
  query {
    posts {
      id
      title
      content
      allowComments
      createdAt
    }
  }
  ```
- Получить пост с комментариями и пагинацией:
  ```
  query {
    post(id: 1) {
      id
      title
      comments(limit: 5, offset: 0) {
        comments {
          id
          content
          branch(limit: 10) {
            comments { id content }
            hasNext
          }
        }
        hasNext
        totalCount
      }
    }
  }
  ```
- Создать пост
  ```
  mutation {
    createPost(input: {
      authorID: 1
      title: "Hello"
      content: "World"
      allowComments: true
    }) {
      id
      title
    }
  }
  ```
- Создать комментарий
  ```
  mutation {
    createComment(input: {
      postID: 1
      content: "Great post!"
      authorID: 2
      parentID: null   
    }) {
      id
      content
      branch {
        comments { id }
      }
    }
  }
  ```
 - Включить/отключить комментарии у поста (может менять только автор):
   ```
   mutation {
      updatePostCommentsAvailability(postID: 1, userID: 1, availability: false) {
        id
        allowComments
      }
    }
   ```
- Подписаться на комментарии к посту:
  ```
  subscription {
    addComment(postID: 1) {
      id
      content
      authorID
      createdAt
    }
  }
  ```

  Структура проекта:
```
  postsys/
├── cmd/                          # точка входа
│   └── server/
│       └── main.go               # запуск сервера, инициализация БД, graceful shutdown
├── internal/                     # внутренний код, не экспортируемый
│   ├── adapter/                  # конвертация между слоями
|       └── adapter.go
│   ├── db/                       # реализация хранилищ
│   │   ├── memdb/                # in‑memory реализация
|   |   |   ├── memdb.go
│   │   |   └── memdb_test.go
│   │   └── pgsql/                # PostgreSQL + sqlc
│   │       ├── pgsql.go          # адаптер для работы с пулом соединений и транзакциями
│   │       └── sqlc/             # сгенерированный sqlc код
│   │           ├── db.go
│   │           ├── models.go     # модели таблиц (Comment, Post)
│   │           └── query.sql.go  # методы, сгенерированные из query.sql
│   ├── entity/                   # доменные сущности
│   |   └── entity.go
│   ├── graph/                    # GraphQL слой
│   │   ├── model/                # сгенерированные модели
│   │   |    └── models_gen.go
│   │   ├── generated.go          # сгенерированный код gqlgen
│   │   ├── resolver.go           # структура Resolver
│   │   ├── context.go            # работа с контекстом 
│   │   └── schema.resolvers.go   # реализация резолверов
│   └── service/                  # бизнес-логика
│       ├── service.go            # интерфейс сервиса
│       ├── service_impl_test.go  # реализация
│       └── service_impl.go       # тесты сервиса (с in‑memory хранилищем)
├── migrations/                   # миграции (goose)
├── sqlc/                         # схемы и запросы для sqlc
|   ├── query.sql                 # SQL-запросы для генерации кода
|   └── schema.sql                # схема БД (одинаковая с миграциями)
├── .env                          # переменные окружения
├── docker-compose.yaml           # запуск БД, миграций и приложения
├── Dockerfile                    # сборка образа приложения
├── go.mod                        # зависимости Go
├── go.sum                        # контрольные суммы зависимостей
├── README.md                     # описание проекта, инструкция по запуску
├── gqlgen.yml                    # конфигурация gqlgen
└── sqlc.yaml                     # конфигурация sqlc
```
