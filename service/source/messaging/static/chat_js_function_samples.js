const getConversationListEvent = "Event.GetConversationList"
const addConversationEvent = "Event.AddConversation"
const addSendMessageEvent = "Event.SendMessage"
const getMessagesList = "Event.GetMessageList"
const readMessage = "Event.ReadMessage"
const unreadCount = "Event.GetUnreadInfo"
const getUnreadMessages = "Event.GetUnreadMessages"

function SendGetConversationListEvent(user){
    var json = JSON.stringify({
        name: getConversationListEvent,
        args : JSON.stringify({
           user_id: user,
        })
    })
   
    ws.send(json)
}


function SendAddConversationEvent(applicationID, userID){
    var json = JSON.stringify({
        name: addConversationEvent,
        args : JSON.stringify({
            user_id: userID,
            application_id: applicationID
        })
    })
   
    ws.send(json)
}

function SendGetMessagesListEvent(sessionChannel,  executorid){
    var json = JSON.stringify({
        name: getMessagesList,
        args: JSON.stringify({
            session_channel: sessionChannel,
            executor_id: executorid
        })
    })

    ws.send(json)
}


function SendSendMessageEvent(sessionChannel, sender, content){
    var json = JSON.stringify({
        name: addSendMessageEvent,
        args : JSON.stringify({
            session_channel:sessionChannel,
            sender_id: sender,
            content: content
        })
    })
   
    ws.send(json)
}

function SendReadMessageEvent(sessionChannel, executor_id, messageid){
    var json = JSON.stringify({
        name: readMessage,
        args : JSON.stringify({
            executor_id: executor_id,
            message_id: messageid
        })
    })
   
    ws.send(json)
}

function SendGetUnreadInfoEvent(userID){
    var json = JSON.stringify({
        name: unreadCount,
        args : JSON.stringify({
            user_id: userID
        })
    })
   
    ws.send(json)
}

function SendGetUnreadMessages(executorID){
    var json = JSON.stringify({
        name: getUnreadMessages,
        args : JSON.stringify({
            executor_id: executorID
        })
    })
   
    ws.send(json)
}

var ws = new WebSocket("ws://localhost:8888/ws?sid=1")   // debug

ws.addEventListener("message", function (data){
    console.log(data.data)
})

