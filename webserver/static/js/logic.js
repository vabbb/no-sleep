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

function getRound(time) {
	$.ajax({
		url: '/round/' + time,
		type: 'POST',
		success: function (response) {
			$("#flow-list").children().each(function (i) {
				this.remove()
			});
			$("#flow-list").append(response);
			flowsFilters.toggle()
		},
		error: function (error) {
			console.log(error);
		}
	});
	
}
// TOGGLE HIGHLITS FOR THE FLOWS
function getActiveFlow(){
	return $("#flow-list").find('.active')
}

function setActiveFlow(id){
	getActiveFlow().removeClass('active')
	$("#flow-list").find('#'+id).addClass('active')
	getFlow(id)
}
// TOGGLE HIGHLITS FOR THE ROUNDS

function getActiveRound(){
	return $("#round-list").find('.active')
}

function setActiveRound(time){
	getActiveRound().removeClass('active')
	$("#round-list").find('#'+time).addClass('active')
	getRound(time)
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
	console.log(event.target.value)
	flowsFilters.activeService = event.target.value
	flowsFilters.toggle()
})