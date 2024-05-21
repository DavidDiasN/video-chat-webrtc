function connectionRequest(name) {
  console.log(name)
}


console.log("meme")
if (window['WebSocket']) {
  const conn = new WebSocket('ws://' + document.location.host + '/videocall/getoffers');
  console.log("Hello")
  let initial = true;
  console.log("we got here") 
  conn.onopen = function () {
    conn.send(inputValue + " " + hash)
  }

  conn.onclose = () => {
    console.log("Connection closing")
  }


  conn.onmessage = evt => {
    if (initial) {
      console.log(evt.data)
    } else {
      console.log("wooo bye")
      conn.send("DONE")
      conn.close()
    }
    
  }
}

