const checkboxHex = document.getElementById('hexdump')
const checkboxFlags = document.getElementById('flagsOnly')
const selectService = document.getElementById('selectService')

// FILTERS OBJECTS
// set the value of the filter and then call the toggle() method to update the changes
var flowFilters = {
	hex : false,

	toggle : function(){
		if (this.hex) {
			$('.blob').removeClass('d-none')
			$('.printableData').addClass('d-none')
		} else {
			$('.blob').addClass('d-none')
			$('.printableData').removeClass('d-none')
		}
	}
}

var flowsFilters = {
	activeService : 'all',
	onlyFlaggedFlows : false,
	onlyFavoriteFlows : false,

	toggle : function(){
		console.log('vengoeseguito')
		$("#flow-list").children().each(function() {
			$(this).removeClass('d-none')
			if (flowsFilters.onlyFlaggedFlows && !$(this).hasClass('hasflag') ){
				if (!$(this).hasClass('d-none'))
					$(this).addClass('d-none')
			}
			if ( flowsFilters.onlyFavoriteFlows && !$(this).hasClass('favourite') ){
				if (!$(this).hasClass('d-none'))
					$(this).addClass('d-none')	
			}
			if ( flowsFilters.activeService != 'all' && !$(this).hasClass(flowsFilters.activeService)){
				if (!$(this).hasClass('d-none'))
					$(this).addClass('d-none')	
			}
		})
	}
}

var roundsFilters = {

}

// UTILITY FUNCTIONS TO FETCH DATA
function getFlow(id) {
	const checkbox = document.getElementById('hexdump')
	$.ajax({
		url: '/flow/' + id + '?hex=' + flowFilters.hex,
		type: 'GET',
		success: function (response) {
			$("#main").empty()
			$("#main").append(response)

			flowFilters.toggle()
		},
		error: function (error) {
			console.log(error);
		}
	});
}

function getRound(time) {
	$.ajax({
		url: '/round/' + time,
		type: 'POST',
		success: function (response) {
			$("#flow-list").empty()
			$("#flow-list").append(response);

			flowsFilters.toggle()
		},
		error: function (error) {
			console.log(error);
		}
	});
	
}

function pwn(flow_id) {
	$.ajax({
		url: '/pwn/' + flow_id,
		type: 'GET',
		success: function (response) {
			$(".modal-body").html("<pre><code>"+
				response+
				"</code></pre>")
			$("#exploit").modal()
		},
		error: function (error) {
			console.log(error);
		}
	});
}

// TOGGLE HIGHLIGHTS FOR THE FLOWS
function getActiveFlow(){
	return $("#flow-list").find('.active')
}

function setActiveFlow(id){
	getActiveFlow().removeClass('active')
	$("#flow-list").find('#'+id).addClass('active')
	getFlow(id)
}

// TOGGLE HIGHLIGHTS FOR THE ROUNDS
function getActiveRound(){
	return $("#round-list").find('.active')
}

function setActiveRound(time){
	getActiveRound().removeClass('active')
	$("#round-list").find('#'+time).addClass('active')
	getRound(time)
}

function shortcut(key){
	switch (key) {
		case 'h':
			checkboxHex.click()
			break;
		default:
			// statements_def
			break;
	}
}

// EVENT LISTENERS
checkboxHex.addEventListener('change', function (event) {
	flowFilters.hex = event.target.checked
	flowFilters.toggle()
})

checkboxFlags.addEventListener('change', function(event) {
	flowsFilters.onlyFlaggedFlows = event.target.checked
	flowsFilters.toggle()
})

selectService.addEventListener('change', function(event) {
	flowsFilters.activeService = event.target.value
	flowsFilters.toggle()
})

document.addEventListener('keydown', function(event) {
	shortcut(event.key)
})

// document.onkeydown = function (e) {
// 	switch (e.key) {
// 		case 'f':
// 			checkboxHex.click()
// 		case 'x':
// 			if (checkboxHex.checked) {
// 				$('#hexdump').prop('checked', false);
// 				onlyShowPrintable();
// 				break;
// 			}
// 			$('#hexdump').prop('checked', true);
// 			onlyShowHexDump();
// 			break;
// 		case 'k':
// 			var curr = $("#flow-list > li.active")
// 			var prev = $("#flow-list > li.active").prevAll().not(".d-none").filter(":first")
// 			if (prev.hasClass("nav-item")) {
// 				curr.removeClass("active")
// 				prev.addClass("active")
// 			}
// 			document.activeElement.blur()
// 			break;
// 		case 'j':
// 			var curr = $("#flow-list > li.active")
// 			if (curr.length == 0) { // select first in list
// 				var first = $("#flow-list > li:not(.d-none)").filter(":first")
// 				first.addClass("active")
// 				break;
// 			}
// 			var next = $("#flow-list > li.active").nextAll().not(".d-none").filter(":first")
// 			if (next.hasClass("nav-item")) {
// 				curr.removeClass("active")
// 				next.addClass("active")
// 			}
// 			document.activeElement.blur()
// 			break;
// 		case 'Enter':
// 			$("#flow-list > li.active").click()
// 			break;
// 		case 'w':
// 			var curr = $("#round-list > li.active")
// 			if (curr.prev().hasClass("nav-item")) {
// 				curr.removeClass("active")
// 				curr.prev().addClass("active")
// 				curr.prev().click()
// 			}
// 			document.activeElement.blur()
// 			break;
// 		case 's':
// 			var curr = $("#round-list > li.active")
// 			if (curr.length == 0) { // select first in list
// 				var first = $("#round-list > li").filter(":first")
// 				first.addClass("active")
// 				first.click()
// 				break;
// 			}
// 			if (curr.next().hasClass("nav-item")) {
// 				curr.removeClass("active")
// 				curr.next().addClass("active")
// 				curr.next().click()
// 			}
// 			document.activeElement.blur()
// 			break;
// 	}
// }