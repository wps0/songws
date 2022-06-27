/*
 * 1.1.0-210325
 */

var ws = new WebSocket('wss://srv1.wieczorekp.pl:2137/ws');
var template = "<span class='time'><img class='currp' src='images/playing.png' /> %%_TIME_BEGIN_%% </span><span class='title'><a href='%%_SONG_URL_%%' target='_blank'>%%_SONG_TITLE_%%</a></span><span class='author'> by <a href='%%_AUTHOR_URL_%%' target='_blank'>%%_AUTHOR_NAME_%%</a></span>";
var monthNames = [
    "stycznia", "lutego", "marca",
    "kwietnia", "maja", "czerwca", "lipca",
    "sierpnia", "wrzeĹnia", "paĹşdziernika",
    "listopada", "grudnia"
  ];

var position = [
  document.getElementById('d1'),
  document.getElementById('d2'),
  document.getElementById('d3')
];

function change_song(obj) {
  var dat = new Date(obj.date * 1000)

	let min  = dat.getMinutes();
	if (min < 10)
		min = "0" + min;
	
	let h  = dat.getHours();
	if (h < 10)
		h = "0" + h;

	position[2].innerHTML = position[1].innerHTML;
	position[1].innerHTML = position[0].innerHTML;
	position[0].innerHTML = template.replace("%%_AUTHOR_URL_%%", "https://www.last.fm/music/" + obj.artist.replace(new RegExp(' ', 'g'), '+'))
								  .replace("%%_AUTHOR_NAME_%%", obj.artist)
								  .replace("%%_SONG_URL_%%", "https://www.last.fm/music/" + obj.artist.replace(new RegExp(' ', 'g'), '+') +"/_/" + obj.title.replace(new RegExp(' ', 'g'), '+'))
								  .replace("%%_SONG_TITLE_%%", obj.title)
								  .replace("%%_TIME_BEGIN_%%", h + ":" + min);
}

ws.onopen = function() {
	console.log("opened");
	position[0].innerHTML = "<h4>Brak piosenek.</h4>";
};

ws.onclose = function(event) {
  console.log(event)
	console.log("closed");
	position[0].innerHTML = "<h4>PoĹÄczenie zostaĹo zamkniÄte!</h4>";
	position[1].innerHTML = "<h4>SprĂłbuj ponownie za jakiĹ czas.</h4>";
	position[2].innerHTML = "";
};

ws.onerror = function(event){
    console.log("Error");
    console.log(event)
}

ws.onmessage = function(event) {
	if (event.data.length == 0) {
		ws.close();
		position[0].innerHTML = "<h3>WystÄpiĹ bĹÄd w trakcie pobierania danych!</h3>";
		position[1].innerHTML = "<h4>SprĂłbuj ponownie za jakiĹ czas.</h4>";
		return;
	}
	console.log(event.data);
	
	var data = JSON.parse(event.data);

  if (data.msg_type == -1){
    return;
  }

  var target = document.getElementsByClassName("currp")[0];
  if (data.msg_type == 0) {
	if (target == undefined)
		return;
	target.classList.add("hidden");
    return
  }
  target.classList.remove("hidden");


  for (let i = data.data.length - 1; i >= 0; i--) {
    const val = data.data[i];
    change_song(val)
  }
	
};