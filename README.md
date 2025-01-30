# kafka_stream_go
Проект по второму модулю

Реализация потоковой обработки сообщений в Kafka + GOKA

Установка:
 1. Перейти в директорию infra
 2. выполнить `sudo docker up -d`
 3. Дождаться запуска kafka
 4. выполнить содержимое topics.txt - это создаст необходимые топики
   

Запуск:
 - для запуска webapi выполнить `go run cmd/webserver/main.go`
 - Для запуска процессоров обработки выполнить `go run cmd/workers/main.go`
 - Для просмотра топиков (kafka UI) и данных можно открыть `http://localhost:8080` 

Работа с приложением:
отправка сообщений пользователю:
`curl -X POST -d '{"from": Yura, "to": "Mike", "content": "Hello, megapups!"}' http://localhost:8081/Yura/send`, отправит сообщение Yura -> Mike

просмотр входящих сообщений пользователя:
 `GET http://localhost:8081/Yura/messages`

заблокировать пользователя:
`curl -X POST -d '{"user": "Yura", "isBlocked": true}'     http://localhost:8081/Mike/block`, после чего сообщения от Yura не будут доставлены пользователю Mike

добавить слово в список стопслов:
`curl -X POST -d '{"key": "ugu", "replacer": "***"}' http://localhost:8081/stopwords/add`, после чего указанное слово будет заменяться на `replacer`

удалить слово из списка стоп-слов:
`curl -X POST -d '{"key": "ugu"}' http://localhost:8081/stopwords/remove`, после чего слова не будут экранироваться

просмотр пользователей которые заблокировали текущего:
 `GET http://localhost:8081/Yura/who_blocked` , выводит кто заблокировал пользователя Yura