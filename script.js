
let socket = new WebSocket("ws://127.0.0.1:8080/ws")
socket.binaryType = "arraybuffer";

document.getElementById("sub").addEventListener("click",StringDataSend)

socket.addEventListener('open', function (event) {
    console.debug("Sucess WS Connect")
});

socket.onmessage = (event) => {
  
  console.debug("receive Data", event.data)
  //var msg = JSON.parse(event.data);
  //console.debug(msg)

}

function ByteDataSend() {

  const data = new Uint8Array(4)
  data[0] = 0x01
  data[1] = 0x02
  data[2] = 0x03
  data[3] = 0x04

  console.debug("send Data", data)
  socket.send(data)

}

function StringDataSend() {

  let message = "test Message"

  console.debug("send Data", message)
  socket.send(message)

}

function JsonDataSend() {

  let msg = {
    data: "test",
    type: "first"
  }
  
  console.debug("send Data", msg)
  socket.send(JSON.stringify(msg))
  
}