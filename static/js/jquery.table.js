	$(function() {
		$("thead").addClass("thead");
		$("thead tr th:has(sortType)").addClass("sortHand");
//		$("tbody tr:nth-child(odd)").addClass("st1");
		$("tbody tr:nth-child(even)").addClass("st2");

		$("tbody tr:has(td:contains('"+myName+"'))").addClass("showme");

//		$("tbody tr").mouseover(function(){
//			$(this).addClass("over");
//		}).mouseout(function(){
//			$(this).removeClass("over");
//		});

	});