<<<<<<< HEAD
function replaceText(title) {
  document.getElementById('CheckName').textContent = title;
  var rawFile = new XMLHttpRequest();
  rawFile.open("GET", "js/output/" + title, false);
  rawFile.onreadystatechange = function () {
    if (rawFile.readyState === 4) {
      if (rawFile.status === 200 || rawFile.status == 0) {
        var allText = rawFile.responseText;
        document.getElementById('CheckOutput').innerHTML = "<pre>" + allText + "</pre>";
      }
    }
  }
  rawFile.send(null);
=======
function replaceText(title, output) {
  document.getElementById('CheckName').textContent = title;

  document.getElementById('CheckOutput').innerHTML = "<pre>" + output + "</pre>";
>>>>>>> d7aa2ba22c5b0ac9b8713bade907a4a6114c68fa
}

$(document).ready(function () {
  $.getJSON("js/current.json",
    function (json) {
      var tr;
      // Append each row to html table
      for (var i = 0; i < json.length; i++) {
        tr = $('<tr/>');
<<<<<<< HEAD
        tr.append("<td><a href='#' onClick='replaceText(\"" + json[i].title + "\");'>" + json[i].title + "</a></td>");
=======
        tr.append("<td><a href='#' onClick='replaceText(\"" + json[i].title + "\"," + json[i].output + ");'>" + json[i].title + "</a></td>");
>>>>>>> d7aa2ba22c5b0ac9b8713bade907a4a6114c68fa
        if (json[i].good == true) {
          tr.append("<td><font color='green'>Success</font></td>");
        }
        else if (json[i].good == false) {
          tr.append("<td><font color='red'>Fail</font></td>");
        }
        tr.append("<td>" + json[i].asof + "</td>");
        $('table').append(tr);
      }
    });
});
