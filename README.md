**Введение**

Чат реализован с использованием протокола web socket. 

На данный момент сервер слушает только **ws** тип, а именно, не защищенное, в дальнейшем планируется добавить возможность использования **wss**

Для успешного получения сокета, необходимо предоставить два параметра в момент установки соединения, а именно:
1.  sid - query string параметр, сокращенно от sender id, а именно уникальный идентификатор участника чата. 
2.  sec-websocket-protocol - http заголовок, который передается как второй параметр, значение должно иметь специальный вид, см. ниже в примере. 

Пример 

`var ws = new WebSocket("ws://localhost:8888/ws?sid=2",["access_token","eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIiLCJpYXQiOm51bGwsImV4cCI6MTYxNDkyMzExNiwiYXVkIjoiIiwic3ViIjoiIiwiYWRtaW5faWQiOiIxIn0.G789lqmV0yUs9sKVTcpid6HK3D4zK98kum-EgWs3yiY"]) `

Параметр sid передается как обычный query string параметр, а значение заголовка должно иметь вид массива, где первый элемент идет access_token, второй это jwt токен, содержащий в себе expiration claim, 
а также один из двух: request_id или admin_id

Использование http заголовка диктуется, отсутствием возможности предоставить Authorization заголовок, в виду архитектурной особенности протокола Web Socket и не возможности нести authorization header.

Использование массива обусловленно тем, что для успешного handshake, сервер должен вернуть один из значений переданных в заголовке, таким образом access_token будет возвращаться.

**Event**

Взаимодействие между клиентской стороной и сервером реализованно основанным на абстракции, Event. Который имеет два свойства, а именно:
1.  name - тип сообщения
2.  args - дочерний json объект, который несет в себе аргументы конкретного типа event.

`{
   "name":"...",
   "args":"..."
}`



**Сообщения**
*  Event.AddConversation
*  Event.GetConversationList
*  Event.GetMessageList
*  Event.SendMessage
*  Event.ReceiveMessage
*  Event.ReadMessage
*  Event.GetUnreadInfo

  
**Event.AddConversation**
1.  user_id - уникальный идентификатор пользователя
2.  application_id - уникальный идентификатор заявки, к нему осуществляется привязка разговора.

Пример: 

`{
   "name":"Event.AddConversation",
   "args":"{\"user_id\":1,\"application_id\":100}"
}`

Response:

Содержит в себе флаг показывающий успешен ли вызов сообщения, тип соощения и результат. 

Пример.

`{
   "name":"Event.AddConversation",
   "ok":true,
   "result":{
      "id":2,
      "session_channel":"100",
      "application_id":100,
      "create_at":"2020-03-05T18:11:50.6982843Z",
      "update_at":"2020-03-05T18:11:50.6982843Z"
   }
}`


**Event.GetConversationList**
1.  user_id - уникальный идентификатор пользователя

Пример:

`{"name":"Event.GetConversationList","args":"{\"user_id\":1}"}`

Response:

Содержит в себе флаг показывающий успешен ли вызов сообщения, тип сообщения а массив объектов conversation

`{"name":"Event.GetConversationList","ok":true,"result":{"conversations":[{"id":1,"session_channel":"1","application_id":1,"create_at":"2020-03-04T16:11:20.751252-05:00","update_at":"2020-03-04T16:11:20.751252-05:00"},{"id":2,"session_channel":"100","application_id":100,"create_at":"2020-03-05T13:11:50.698284-05:00","update_at":"2020-03-05T13:11:50.698284-05:00"}]}}`


**Event.GetMessageList** 

Ивент отвечает за возвращение списка сообщений для какого то конректного объекта conversation

1.  session_channel - идентификатор объекта conversation полученный после вызова GetMessageList
2.  executor_id - идентификатор пользователя от имени которого происходит вызов ивента

Пример:

`{
   "name":"Event.GetMessageList",
   "args":"{\"session_channel\":\"100\",\"executor_id\":1}"
}`

Response:

Содержит в себе флаг описывающий успешен ли вызов сообщения, тип сообщения, а также массив объектов message
  
Пример:

`{
   "name":"Event.GetMessageList",
   "ok":true,
   "result":{
      "Messages":[
         {
            "id":1,
            "session_channel":"100",
            "content":"Hello World",
            "sender_id":1,
            "status":0,
            "create_at":"2020-03-05T13:20:15.948653-05:00",
            "update_at":"2020-03-05T13:20:15.948653-05:00"
         }
      ]
   }
}

**Event.SendMessage**

Ивент отвечает за отсылку сообщения в контексте conversation

1.  session_channel - идентификатор объекта conversation полученный после вызова SendMessage
2.  sender_id - идентификатор пользователя который от имени которого происходит отсылка сообщения
3.  content - контент сообщения
4.  content_type - опциональный параметр, который указывает тип сообщения. Может быть 1 или 2. 1 соответствует тексту, 2 - изображение. 
                   В случае если никакого параметра не отсылается или отсылается параметр отличный от 1 и 2, контент тип текст будет использован
   

Пример

`{
   "name":"Event.SendMessage",
   "args":"{\"session_channel\":\"10\",\"sender_id\":1,\"content\":\"Hey\"}"
}`


Response

Содержит в себе флаг описывающий успешен ли вызов сообщения, тип сообщения, а также объект message

Пример

`{
   "name":"Event.SendMessage",
   "ok":true,
   "result":{
      "message":{
         "id":10,
         "session_channel":"10",
         "content":"Hey",
         "sender_id":1,
         "status":0,
         "create_at":"2020-03-06T03:29:29.321142339Z",
         "update_at":"2020-03-06T03:29:29.321142998Z"
      }
   }
}`


**Event.ReceiveMessage**

Ивент уведомляет о поступлении нового сообщения, в случае если оппонент присоединен к беседе в данный момент. Этот ивент отсылается с сервера.

Пример

`{
   "name":"Event.ReceiveMessage",
   "ok":true,
   "result":{
      "message":{
         "id":11,
         "session_channel":"10",
         "content":"hey",
         "sender_id":1,
         "status":0,
         "create_at":"2020-03-06T03:41:13.944327781Z",
         "update_at":"2020-03-06T03:41:13.944328092Z"
      }
   }
}


**Event.ReadMessage**`

Ивент осуществляет процедуру маркировки сообщения как прочитанно.

1.  session_channel - идентификатор объекта conversation полученный после вызова ReceiveMessage
2.  executor_id - идентификатор пользователя от имени которого происходит вызов ивента
3.  messagE_id - идентификатор сообщения который необходимо отметить как прочитанный

Пример 

`{
   "name":"Event.ReadMessage",
   "args":"{\"session_channel\":\"10\",\"executor_id\":2,\"message_id\":11}"
}`

Response

Содержит в себе флаг описывающий успешен ли вызов сообщения, тип сообщения.

Пример

`{
   "name":"Event.ReadMessage",
   "ok":true,
   "result":null
}`


