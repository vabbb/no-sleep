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
	if (serviceActived == "all") $('li.flow').not('li.flow.hasflag').addClass('d-none')
	else $('li.flow.' + serviceActived).not('li.flow.hasflag.' + serviceActived).addClass('d-none')
}

function undoShowFlagsOnly() {
	if (serviceActived == "all") $('li.flow').removeClass('d-none')
	else $('li.flow.' + serviceActived).removeClass('d-none')
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
		showFlagsOnly();
	} else {
		undoShowFlagsOnly();
	}
})

selectService.addEventListener('change', (event) => {
	service = event.target.value
	if (service == "all") {
		if (!checkboxFlags.checked) $('li.flow').removeClass('d-none')
		else {
			$('li.flow').removeClass("d-none")
			$('li.flow').not('li.flow.hasflag').addClass('d-none')
		}
	} else {
		if (!checkboxFlags.checked) {
			$('li.flow').removeClass("d-none")
			$('li.flow').not('li.flow.' + service).addClass("d-none")
		} else {
			$('li.flow').removeClass("d-none")
			$('li.flow').not('li.flow.hasflag.' + service).addClass('d-none')
		}
	}
})

/* *****************************************************************************************************************
   ***************************************************************************************************************** */

var request = self.indexedDB.open('EXAMPLE_DB', 1);
var db;
var lastUpdate = 0

request.onsuccess = function (event) {
	console.log('[onsuccess]', request.result);
	db = event.target.result;
};

request.onerror = function (event) {
	console.log('[onerror]', request.error);
};

request.onupgradeneeded = function (event) {
	var db = event.target.result;
	var store = db.createObjectStore('flows', { keyPath: '_id' });
	store.createIndex('time', 'time', { unique: false });
};

function add_flows_to_db(flows) {
	var transaction = db.transaction('flows', 'readwrite');
	transaction.onsuccess = function (event) {
		console.log('[Transaction] ALL DONE!');
	};
	var flowsStore = transaction.objectStore('flows');

	flows.forEach(function (flow) {
		temp = { _id: flow['_id']["$oid"], time: flow['time'], flow: flow }
		//time_l.push(flow['time'])
		var db_op_req = flowsStore.add(temp);
	});
}

function get_new_flows() {
	$.ajax({
		url: '/flows/' + lastUpdate,
		type: 'GET',
		success: function (response) {

			flows = JSON.parse(response);
			//time_l = []
			add_flows_to_db(flows)

			//lastUpdate = Math.max(...time_l)
		},
		error: function (error) {
			console.log(error);
		}
	});
}

get_new_flows()
//setInterval(get_new_flows, 1000)