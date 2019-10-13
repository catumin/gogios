function replaceText(title) {
    document.getElementById('CheckName').textContent = title;
    var rawFile = new XMLHttpRequest();
    rawFile.open("GET", "js/output/" + title, false);
    rawFile.onreadystatechange = function () {
        if (rawFile.readyState === 4) {
            if (rawFile.status === 200 || rawFile.status == 0) {
                var allText = rawFile.responseText;
                document.getElementById('CheckOutput').innerHTML = "<xmp>" + allText + "</xmp>";
            }
        }
    }
    rawFile.send(null);
}
