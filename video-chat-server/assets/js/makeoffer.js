const servers = {
  iceServers: [
    {
      urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302']
    },
  ],
  iceCandidiatePoolSize: 10,
};
let conn;
let peerConn = new RTCPeerConnection(servers);
let localStream = null;
let remoteStream = null;

const protocolSep = "/:|:/"
const webcamButton = document.getElementById("enable-webcam");
const webcamVideo = document.getElementById("local-video");
const remoteVideo = document.getElementById("remote-video");
const incomingOfferMap = new Map();


function submitOfferName() {
  let inputButton = document.getElementById("user-input");
  let inputValue = inputButton.value;
  let hash = "";

  fetch("http://127.0.0.1:4009/videocall/OfferValidation", {
    method: "POST",
    body: inputValue,
  })
  .then((response) => response.text())
  .then((text) => {
    if (text === "NO") {
        console.log("it said no")
        return
        // have a popup say that name is taken
      } else {
        console.log(text);
        inputButton.disabled = true;
        hash = text;

        if (window['WebSocket']) {
          conn = new WebSocket('ws://' + document.location.host + '/videocall/MakeOffer/ws');
          let initial = true;
          let answerList = document.getElementById("answer-list")
          console.log("ws started")

          conn.onopen = function () {
            conn.send(inputValue + " " + hash);
          }

          conn.onclose = function () {
            console.log("Connection closing");
          }

          conn.onmessage = evt => {
            let messageUnwrapped = evt.data.split("/:|:/");
            console.log("MESSAGE Type: " + messageUnwrapped[0]);

            if (initial) {
              if (evt.data === "Invalid Hash") {
                //whatever
                console.log("STOPSTOPSTOPSOTP")
                offerList.innerText = "STOPSTOPSTOPSTOP"
                return
              } else {

                initial = false;
              }
            }            

            if (messageUnwrapped[0] === "Offer") {
              answerList.innerHTML += messageUnwrapped[1];
              incomingOfferMap.set(answerList.innerText, messageUnwrapped[2]);

            } else if (evt.data === "DONE") {
              console.log("wooo bye");
              conn.send("DONE");
              conn.close();
            } else if (messageUnwrapped[0] === "Ice") {

              console.log("Ice route")
              if (messageUnwrapped[1] !== connectionPeerName || connectionPeerName === "" ) {
                console.log("No way jose")
                return
              }

              const candidate = new RTCIceCandidate(JSON.parse(messageUnwrapped[2]));
              peerConn.addIceCandidate(candidate);
            }
          }
        }
      }
    });
}

async function clickName(targetName) {
  console.log("Start")

  peerConn.onicecandidate = event => {
    if (event.candidate != null) {
      conn.send("Ice"+ protocolSep + targetName + protocolSep + JSON.stringify(event.candidate));
    }
  }

  let offerObject = JSON.parse(incomingOfferMap.get(targetName));
  await peerConn.setRemoteDescription(new RTCSessionDescription(offerObject));
  const answerObject = await peerConn.createAnswer();
  await peerConn.setLocalDescription(answerObject);

  console.log(offerObject);
  if (conn !== null) {
    conn.send("Answer" + protocolSep + targetName + protocolSep + JSON.stringify(answerObject));
  } 
  console.log("End");
} 

webcamButton.onclick = async () => {
  // you need to remove the audio track from the local stream because it causes an echo
  localStream = await navigator.mediaDevices.getUserMedia({video: true, audio: true});
  remoteStream = new MediaStream();

  localStream.getTracks().forEach((track) => {
    peerConn.addTrack(track, localStream);
  });

  peerConn.ontrack = event => {
    event.streams[0].getTracks().forEach(track => {
      remoteStream.addTrack(track);
    });
  };

  webcamVideo.srcObject = localStream;
  remoteVideo.srcObject = remoteStream;

};

