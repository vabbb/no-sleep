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
			$(".modal-body").html("<pre><code>"+
				response+"</code></pre>");
			$("#exploit").modal();
		},
		error: function (error) {
			console.log(error);
		}
	});
}

function copyExploitToClipboard() {
	var $temp = $("<textarea>");
	$(".modal-body").append($temp);
	$temp.text($(".modal-body").text()).select();
	document.execCommand("copy");
	$temp.remove();
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
		case 'k':
			var curr = $("#flow-list > li.active")
			var prev = $("#flow-list > li.active").prevAll().not(".d-none").filter(":first")
			if (prev.hasClass("nav-item")) {
				curr.removeClass("active")
				prev.addClass("active")
			}
			document.activeElement.blur()
			break;
		case 'j':
			var curr = $("#flow-list > li.active")
			if (curr.length == 0) { // select first in list
				var first = $("#flow-list > li:not(.d-none)").filter(":first")
				first.addClass("active")
				break;
			}
			var next = $("#flow-list > li.active").nextAll().not(".d-none").filter(":first")
			if (next.hasClass("nav-item")) {
				curr.removeClass("active")
				next.addClass("active")
			}
			document.activeElement.blur()
			break;
		case 'Enter':
			$("#flow-list > li.active").click()
			break;
		case 'w':
			var curr = $("#round-list > li.active")
			if (curr.prev().hasClass("nav-item")) {
				curr.removeClass("active")
				curr.prev().addClass("active")
				curr.prev().click()
			}
			document.activeElement.blur()
			break;
		case 's':
			var curr = $("#round-list > li.active")
			if (curr.length == 0) { // select first in list
				var first = $("#round-list > li").filter(":first")
				first.addClass("active")
				first.click()
				break;
			}
			if (curr.next().hasClass("nav-item")) {
				curr.removeClass("active")
				curr.next().addClass("active")
				curr.next().click()
			}
			document.activeElement.blur()
			break;
	}
}