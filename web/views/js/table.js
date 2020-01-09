$(document).ready(function () {
  $.getJSON("js/current.json",
    function (json) {
      var tr;
      // Append each row to html table
      for (var i = 0; i < json.length; i++) {
        tr = $('<tr/>');
        tr.append("<td><a href='#' onClick='replaceText(\"" + json[i].title + "\");'>" + json[i].title + "</a></td>");
        if (json[i].status == "Success") {
          tr.append("<td><font color='green'>Success</font></td>");
        }
        else if (json[i].status == "Failed") {
          tr.append("<td><font color='red'>Failed</font></td>");
        }
        else if (json[i].status == "Timed Out") {
          tr.append("<td><font color='orange'>Timed Out</font></td>");
        }

        ratio = Math.round(json[i].good_count / json[i].total_count * 100);

        tr.append("<td>" + ratio + "% Uptime</td>");
        tr.append("<td>" + json[i].asof + "</td>");
        $('table').append(tr);
      }
    });
});
