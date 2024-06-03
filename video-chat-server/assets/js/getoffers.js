
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
let connectionPeerName = "";

const protocolSep = "/:|:/"
const webcamButton = document.getElementById("enable-webcam");
const webcamVideo = document.getElementById("local-video");
const remoteVideo = document.getElementById("remote-video");

function submitAnswerName() {
  let inputButton = document.getElementById("user-input");
  let inputValue = inputButton.value;
  let hash = "";
  fetch("http://127.0.0.1:4009/videocall/AnswerValidation", {
    method: "POST",
    body: inputValue,
  })
  .then((response) => response.text())
  .then((text) => {
    if (text === "NO") {
        console.log("it said no")
        return
        // have a popup say that name is taken or something
      } else {
        console.log(text);
        inputButton.disabled = true;
        hash = text;

        if (window['WebSocket']) {
          conn = new WebSocket('ws://' + document.location.host + '/videocall/MakeAnswer/ws');
          let initial = true;
          let offerList = document.getElementById("offer-list")

          conn.onopen = function () {
            conn.send(inputValue + " " + hash)
          };

          conn.onclose = function () {
            console.log("Connection closing")
          };

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

            if (messageUnwrapped[0] === "DONE") {
              console.log("wooo bye")
              conn.send("DONE")
              conn.close()
              return
            } else if (messageUnwrapped[0] === "Offers") {
              offerList.innerHTML = messageUnwrapped[1] 
            } else if (messageUnwrapped[0] === "Answer") {
              console.log(messageUnwrapped[2]);
              connectionPeerName = messageUnwrapped[1];
              const answerObject = new RTCSessionDescription(JSON.parse(messageUnwrapped[2]))
              peerConn.setRemoteDescription(answerObject)

            } else if (messageUnwrapped[0] === "Ice") {
              console.log("Ice route")
              peerConn.addIceCandidate(JSON.parse(messageUnwrapped[2]));
            }

          };


        }
      }

    });
}

async function clickName(targetName) {
  
  
  const offerDescription = await peerConn.createOffer();
  await peerConn.setLocalDescription(offerDescription);
  
  const offer = {
    sdp: offerDescription.sdp,
    type: offerDescription.type,
  };

  if (conn !== null) {
    conn.send("Offer" + protocolSep + targetName + protocolSep + JSON.stringify(offer));
  } 


  peerConn.onicecandidate = event => {
    if (event.candidate) {
      conn.send("Ice"+ protocolSep + targetName + protocolSep + JSON.stringify(event.candidate));
    }
  }

} 


webcamButton.onclick = async () => {
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

peerConn.addEventListener('connectionstatechange', _ => {
    if (peerConn.connectionState === 'connected') {
      console.log("Peer connection established")
    }
});

