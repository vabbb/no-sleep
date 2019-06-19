function get_flow(id) {
	const checkbox = document.getElementById('hexdump')
	$.ajax({
		url: '/flow/' + id + '?hex=' + checkbox.checked,
		type: 'GET',
		success: function (response) {
			$("#main").children().each(function (i) {
				this.remove()
			})
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
			$("#flow-list").children().each(function (i) {
				this.remove()
			});
			// console.log(response);
			$("#flow-list").append(response);
		},
		error: function (error) {
			console.log(error);
		}
	});
}
/*
function update_starred() {
	$(".list-group-star").children().each(function (i) {
		this.remove()
	})
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

function showFlagsOnly() {
	serviceActived = selectService.value
	$('li.flow').removeClass("d-none")
	if (serviceActived == "all") $('li.flow').not('li.flow.hasflag').addClass('d-none')
	else $('li.flow').not('li.flow.hasflag.' + serviceActived).addClass('d-none')
}

function undoShowFlagsOnly() {
	serviceActived = selectService.value
	if (serviceActived == "all") $('li.flow').removeClass('d-none')
	else {
		$('li.flow').removeClass('d-none')
		$('li.flow').not('li.flow.' + serviceActived).addClass('d-none')
	}
}

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

const checkboxHex = document.getElementById('hexdump')
const checkboxFlags = document.getElementById('flagsOnly')
const selectService = document.getElementById('selectService')

function onlyShowPrintable() {
	$('.blob').addClass('d-none')
	$('.printableData').removeClass('d-none')
}

function onlyShowHexDump() {
	$('.blob').removeClass('d-none')
	$('.printableData').addClass('d-none')
}

checkboxHex.addEventListener('change', (event) => {
	if (event.target.checked) {
		onlyShowHexDump();
	} else {
		onlyShowPrintable();
	}
})

checkboxFlags.addEventListener('change', (event) => {
	if (event.target.checked) {
		showFlagsOnly();
	} else {
		undoShowFlagsOnly();
	}
})

selectService.addEventListener('change', (event) => {
	if (checkboxFlags.checked) {
		showFlagsOnly()
	} else {
		undoShowFlagsOnly()
	}
})

document.onkeydown = function (e) {
	switch (e.key) {
		case 'f':
			if (checkboxFlags.checked) {
				$('#flagsOnly').prop('checked', false);
				undoShowFlagsOnly();
				break
			}
			$('#flagsOnly').prop('checked', true);
			showFlagsOnly();
			break;
		case 'x':
			if (checkboxHex.checked) {
				$('#hexdump').prop('checked', false);
				onlyShowPrintable();
				break;
			}
			$('#hexdump').prop('checked', true);
			onlyShowHexDump();
			break;
		case 'j':
			var curr = $("#flow-list > li.active")
			curr.removeClass("active")
			curr.prev().addClass("active")
			break;
		case 'k':
			var curr = $("#flow-list > li.active")
			curr.removeClass("active")
			curr.next().addClass("active")
			break;
		case 'Enter':
			$("#flow-list > li.active").click()
			break;
		case 'w':
			var curr = $("#round-list > li.active")
			curr.removeClass("active")
			curr.prev().addClass("active")
			curr.prev().click()
			break;
		case 's':
			var curr = $("#round-list > li.active")
			curr.removeClass("active")
			curr.next().addClass("active")
			curr.next().click()
			break;
	}
}