var app = new Vue({
    el: '#app',
    data: {
        ws: null,
        serverUrl: "ws://localhost:8080/ws",
        roomInput: null,
        rooms: [],
        user: {
            name: ""
        },
        users: []
    },
    mounted: function () {

    },
    methods: {
        connect() {
            this.connectToWebsocket();
        },
        connectToWebsocket() {
            this.ws = new WebSocket(this.serverUrl + "?name=" + this.user.name);
            this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
            this.ws.addEventListener('message', (event) => { this.handleNewMessage(event) });
        },
        onWebsocketOpen() {
            console.log("connected to WS!");
        },

        handleNewMessage(event) {
            let data = event.data;
            data = data.split(/\r?\n/);

            for (let i = 0; i < data.length; i++) {
                let msg = JSON.parse(data[i]);
                console.log(msg)
                switch (msg.action) {
                    case "TextMessage":
                        this.handleChatMessage(msg);
                        break;
                    case "RoomJoin":
                        this.handleUserJoined(msg);
                        break;
                    case "RoomLeave":
                        this.handleUserLeft(msg);
                        break;
                    case "JoinedRoom":
                        this.handleRoomJoined(msg);
                        break;
                    default:
                        break;
                }

            }
        },
        handleChatMessage(msg) {
            const room = this.findRoom(msg.target);
            if (typeof room !== "undefined") {
                console.log("here")
                room.messages.push(msg);
            }
        },
        handleUserJoined(msg) {
            this.users.push(msg.sender);
            const room = this.findRoom(msg.target);

            if (typeof room !== "undefined") {
                console.log("here")
                room.messages.push(msg);
            }
        },
        handleUserLeft(msg) {
            for (let i = 0; i < this.users.length; i++) {
                if (this.users[i].id == msg.sender.id) {
                    this.users.splice(i, 1);
                }
            }
        },
        handleRoomJoined(msg) {
            let room = {};
            room.name = msg.body
            room.id = msg.target
            room["messages"] = [];
            this.rooms.push(room);
        },
        sendMessage(room) {
            if (room.newMessage !== "") {
                this.ws.send(JSON.stringify({
                    action: 'TextMessage',
                    body: room.newMessage.trim(),
                    target: room.id
                }));
                room.newMessage = "";
            }
        },
        findRoom(roomId) {
            for (let i = 0; i < this.rooms.length; i++) {
                if (this.rooms[i].id === roomId) {
                    return this.rooms[i];
                }
            }
        },
        joinRoom() {
            this.ws.send(JSON.stringify({ action: 'JoinRoom', body: this.roomInput }));
            this.roomInput = "";
        },
        leaveRoom(room) {
            this.ws.send(JSON.stringify({ action: 'LeaveRoom', body: room.id }));

            for (let i = 0; i < this.rooms.length; i++) {
                if (this.rooms[i].id === room.id) {
                    this.rooms.splice(i, 1);
                    break;
                }
            }
        },
        joinPrivateRoom(room) {
            this.ws.send(JSON.stringify({ action: 'join-room-private', message: room.id }));
        }
    }
})