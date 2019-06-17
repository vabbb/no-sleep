function get_flow(id) {
    const checkbox = document.getElementById('hexdump')
    $.ajax({
        url: '/flow/' + id + '?hex=' + checkbox.checked,
        type: 'GET',
        success: function (response) {
            $("#main").children().each(function (i) { this.remove() })
            $("#main").append(response)
        },
        error: function (error) {
            console.log(error);
        }
    });
}

function get_round(time) {
    $.ajax({
        url: '/round/' + time,
        type: 'POST',
        success: function (response) {
            $("#flow-list").children().each(function (i) { this.remove() });
            console.log(response);
            $("#flow-list").append(response);
        },
        error: function (error) {
            console.log(error);
        }
    });
}
/*
function update_starred() {
    $(".list-group-star").children().each(function (i) { this.remove() })
    $.ajax({
        url: '/starred',
        type: 'POST',
        success: function (response) {
            $(".list-group-star").append(response)
        },
        error: function (error) {
            console.log(error);
        }
    });
}

function change_star(icon, id) {
    var sel
    if (icon.className.match(/far/)) {
        icon.className = icon.className.replace(/far/g, 'fas')
        sel = 'true'
    } else {
        icon.className = icon.className.replace(/fas/g, 'far')
        sel = 'false'
    }
    $.ajax({
        url: '/star/' + id + '/' + sel,
        type: 'POST',
        success: function (response) {
            console.log(response);
        },
        error: function (error) {
            console.log(error);
        }
    });
    update_starred()
}
*/
function deactivate_all() {
    $("#flow-list").children().each(function (i) {
        this.className = this.className.replace(/ active/, '')
    })
}

function activate(o, id) {
    deactivate_all()
    if (!o.className.match(/active/)) {
        o.className += ' active'
    }
    get_flow(id)
}

function deactivate_all_rounds() {
    $("#round-list").children().each(function (i) {
        this.className = this.className.replace(/ active/, '')
    })
}

function activate_round(o, time) {
    deactivate_all_rounds()
    if (!o.className.match(/active/)) {
        o.className += ' active'
    }
    get_round(time)
}
/*
function pwn(flow_id) {
    $.ajax({
        url: '/pwn/' + flow_id,
        type: 'GET',
        success: function (response) {
            alert(response);
        },
        error: function (error) {
            console.log(error);
        }
    });
}
*/
const checkboxHex = document.getElementById('hexdump')
const checkboxFlags = document.getElementById('flagsOnly')
const selectService = document.getElementById('selectService')

checkboxHex.addEventListener('change', (event) => {
  if (event.target.checked) {
    $('.blob').removeClass('d-none')
    $('.printableData').addClass('d-none')
  } else {
    $('.blob').addClass('d-none')
    $('.printableData').removeClass('d-none')
  }
})

checkboxFlags.addEventListener('change', (event) => {
	serviceActived = selectService.value
  if (event.target.checked) {
		if(serviceActived == "all") $('li.flow').not('li.flow.hasflag').addClass('d-none')
		else $('li.flow.'+serviceActived).not('li.flow.hasflag.'+serviceActived).addClass('d-none')
  } else {
		if(serviceActived == "all") $('li.flow').removeClass('d-none')
		else $('li.flow.'+serviceActived).removeClass('d-none')		
  }
})

selectService.addEventListener('change', (event) => {
    service = event.target.value
    if (service == "all") {
			if(!checkboxFlags.checked) $('li.flow').removeClass('d-none')
			else {
				$('li.flow').removeClass("d-none")
				$('li.flow').not('li.flow.hasflag').addClass('d-none')
			}
    } else {
			if(!checkboxFlags.checked){
				$('li.flow').removeClass("d-none")
				$('li.flow').not('li.flow.'+service).addClass("d-none")
			} 
			else {
				$('li.flow').removeClass("d-none")
				$('li.flow').not('li.flow.hasflag.'+service).addClass('d-none')
			}
    }
  })