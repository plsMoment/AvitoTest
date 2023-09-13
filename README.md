# Тестовое задание для стажёра Backend
# Сервис динамического сегментирования пользователей

Для запуска сервера
```bash
make run
```
Для ребилда контейнера
```bash
make build
```

Пример создания сегмента (поле slug должно быть строкой, т.е одним запросом создается один сегмент)
```bash
curl -X POST -H "Content-Type: application/json" -d '{"slug": "example"}' http://localhost:8080/segment
curl -X POST -H "Content-Type: application/json" -d '{"slug": "xmpl"}' http://localhost:8080/segment
curl -X POST -H "Content-Type: application/json" -d '{"slug": "ex"}' http://localhost:8080/segment
```

Пример удаления сегмента (поле slug должно быть строкой, т.е одним запросом удаляется один сегмент)
```bash
curl -X DELETE http://localhost:8080/segment/xmpl
```

Пример добавления и удаления сегментов пользователя
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"add_slugs": ["example", "ex"], "delete_slugs": []}' http://localhost:8080/segments/user/a6c87a2a-de4c-4f85-8589-34736f6ea1e1
```

Пример получения активных сегментов пользователя
```bash
curl -X GET http://localhost:8080/segments/user/a6c87a2a-de4c-4f85-8589-34736f6ea1e1
```