$(function() {
    console.log($( ".unixtime" ).length);
    if ( $( ".unixtime" ).length ) {
        $(".unixtime").each(function () {
            var time_val =$(this).text();
            if (time_val) {
                var time = Number($(this).text() + '000');
                /*var d = new Date(time);
                $(this).text(d);*/
                var d = new Date();
                d.setTime(time);
                $(this).text(d.toLocaleString());
            }
        });
    }
});