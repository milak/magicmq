var Instance = function (aHost,aPort) {
    this.host = aHost; 
    this.port = aPort;
    this.topics = new Array();
}
Instance.prototype.addTopic(aTopic){
	this.topics.push(topic);
}
var currentInstance = null;
function toInstance(aName){
	var index = aName.indexOf(":");
	return {
		host : aName.substring(0, index),
		port : aName.substring(index + 1)
	};
}
var instances = new Array();
var store = {
	isEmpty : function () {
		var instance_list = getCookie("instance_list")
		return instance_list == "";
	},
	containsInstance : function (aInstance){
			var instance_list = this.getInstances();
			for (var i = 0; i < instance_list.length; i++) {
				var instance = instance_list[i];
				if ((instance.host == aInstance.host) && (instance.port == aInstance.port)) {
					return true;
				}
			}
			return false;
	},
	addInstance : function (aInstance) {
		if (this.containsInstance(aInstance)) {
			return;
		}
		var instance_list = getCookie("instance_list")
		if (instance_list != "") {
			instance_list += ",";
		}
		instance_list += aInstance.host + ":" + aInstance.port;
		setCookie("instance_list", instance_list, 10);
	},
	getInstances : function() {
		var result = new Array();
		var instance_list = getCookie("instance_list");
		if (instance_list.length == 0){
			return result;
		}
		var pos = instance_list.indexOf(",");
		if (pos == -1) {
			result.push(toInstance(instance_list));
		} else {
			var previous = 0;
			while (pos != -1) {
				result.push(toInstance(instance_list.substring(previous, pos)));
				previous = pos + 1;
				pos = instance_list.indexOf(",", pos + 1);
			}
			result.push(toInstance(instance_list.substring(previous)));
		}
		return result;
	}
};


function addInstancePanel(data) {
	$('#accordion').append(
			'<h3>' + data.Name + '</h3><div><p>Version : ' + data.Version
					+ '<button>Remove instance</button></p></div>').accordion(
			"refresh");
}

function loadInstance(instance, addToList) {
	var url = "http://" + instance.host + ":" + instance.port + "/info";
	$.ajax({
		url : url,
		success : function(data) {
			addInstancePanel(data);
			if (addToList) {
				store.addInstance(new Instance(data.Host, data.Port));
			}
		},
		error : function(jqXHR, textStatus, errorThrown) {
			alert("Error " + textStatus + " " + errorThrown);
		},
		dataType : "json"
	});
}
function reloadInstancesInStore() {
	var instance_list = store.getInstances();
	for (var i = 0; i < instance_list.length; i++) {
		var instance = instance_list[i];
		loadInstance(instance, false);
		if (i == 0){
			loadInformation(instance);
		}
	}
}
function showAddInstance() {
	var dialog = $("#dialog-form").dialog({
		autoOpen : false,
		height : 200,
		width : 350,
		modal : true,
		buttons : {
			"Add" : function() {
				var host = $("#addinstance-host").val();
				var port = $("#addinstance-port").val();
				loadInstance(new Instance(host, port), true);
				dialog.dialog("close");
			},
			Cancel : function() {
				dialog.dialog("close");
			}
		},
		close : function() {
			// form[ 0 ].reset();
			// allFields.removeClass( "ui-state-error" );
		}
	});
	dialog.dialog("open");
}
$(function() {
	$("#accordion").accordion({
		activate : function(event, ui) {
			loadInformation(toInstance(ui.newHeader.text()));
		}
	});
	$("#tabs").tabs();
	if (!store.isEmpty()){
		reloadInstancesInStore();
	} else {
		$.ajax({
			url : "/info",
			success : function(data) {
				addInstancePanel(data);
				var instance = new Instance(data.Host,data.Port);
				store.addInstance(instance);
				loadInformation(instance);
			},
			error : function(jqXHR, textStatus, errorThrown) {
				alert("Error " + textStatus + " " + errorThrown);
			},
			dataType : "jsonp"
		});
	}
});
function loadInformation(instance) {
	currentInstance = instance;
	$.ajax({
		url : "http://" + instance.host+":"+instance.port + "/topic",
		success : function(data) {
			var topic_list = "<table>";
			topic_list += "<thead><tr><th>Name</th><th>Type</th></tr></thead><tbody>";
			for (var t = 0 ; t < data.length; t++){
				topic_list+= "<tr>";
				topic_list += "<td>"+data[t].Name+"</td>";
				topic_list += "<td>"+data[t].Type+"</td>";
				topic_list+= "</tr>";
			}
			topic_list+= "</tbody></table>";
			$("#topic-list").html(topic_list);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			alert("Error " + textStatus + " " + errorThrown);
		},
		dataType : "jsonp"
	})
}
function result(data) {
	alert("received " + data)
}