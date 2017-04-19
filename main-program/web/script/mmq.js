var Instance = function(aHost, aPort) {
	this.host = aHost;
	this.port = aPort;
	this.topics = new Array();
}
Instance.prototype.addTopic = function(aTopic) {
	this.topics.push(topic);
}
Instance.prototype.toString = function() {
	return this.host + ":" + this.port;
}
var currentInstance = null;
var currentTopic = null;
function toInstance(aName) {
	var index = aName.indexOf(":");
	return {
		host : aName.substring(0, index),
		port : aName.substring(index + 1)
	};
}
var instances = new Array();
var store = {
	isEmpty : function() {
		var instance_list = getCookie("instance_list")
		return instance_list == "";
	},
	containsInstance : function(aInstance) {
		var instance_list = this.getInstances();
		for (var i = 0; i < instance_list.length; i++) {
			var instance = instance_list[i];
			if ((instance.host == aInstance.host)
					&& (instance.port == aInstance.port)) {
				return true;
			}
		}
		return false;
	},
	addInstance : function(aInstance) {
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
		if (instance_list.length == 0) {
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
function addInstancePanel(instance, error) {
	var html = '<div><p>';
	if (error != null) {
		html += 'Unable to connect ' + error + '<br/>';
		// html += '<button
		// onclick="loadInstance(\''+instance.toString()+'\',false)"></button><br/>';
	} else {
		html += 'Version : ' + instance.version + '<br/>';
	}
	html += '<a href="#" class="button">Remove instance</a>';
	$('#accordion').append(
			'<h3>' + instance.toString() + '</h3>' + html + '</p></div>')
			.accordion("refresh");
}
function loadInstance(instance, addToList) {
	var url = "http://" + instance.host + ":" + instance.port + "/info";
	$.ajax({
		url : url,
		success : function(data) {
			var instance = new Instance(data.Host, data.Port);
			instance.version = data.Version;
			addInstancePanel(instance, null);
			if (addToList) {
				store.addInstance(instance);
			}
		},
		error : function(jqXHR, textStatus, errorThrown) {
			addInstancePanel(instance, "Unreachable " + errorThrown);
		},
		dataType : "json"
	});
}
function reloadInstancesInStore() {
	var instance_list = store.getInstances();
	for (var i = 0; i < instance_list.length; i++) {
		var instance = instance_list[i];
		loadInstance(instance, false);
		if (i == 0) {
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
function loadTopic(aTopicName) {
	$("#form-topic-button").prop('disabled', true);
	currentTopic = null;
	$.ajax({
		url : "http://" + currentInstance.host + ":" + currentInstance.port
				+ "/topic/" + aTopicName,
		success : function(data) {
			currentTopic = data.Name;
			$("#tabs").tabs("option", "active", 2);
			$("#form-topic-title").html(data.Name);
			$("#form-topic-button").prop('disabled', false);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			alert("Error " + textStatus + " " + errorThrown);
		},
		dataType : "jsonp"
	})
}
function loadInformation(instance) {
	$("#form-topic-title").html("");
	$("#form-topic-button").prop('disabled', true);
	$("#form-create-item-submit").prop('disabled', true);
	currentInstance = instance;
	$.ajax({
		url : "http://" + instance.host + ":" + instance.port + "/topic",
		success : function(data) {
			var topic_list = "";
			var formCreateItemTopicList = "";
			for (var t = 0; t < data.length; t++) {
				var topic = data[t];
				if (topic.Type == "SIMPLE") {
					formCreateItemTopicList += "<tr><td><input type='checkbox' name='topic' value='" + topic.Name + "'/></td><td>" + topic.Name + "</td></tr>";
				}
				topic_list += "<tr>";
				topic_list += "<td><a href='#' onclick='loadTopic(\"" + topic.Name + "\")'>" + topic.Name + "</a></td>";
				topic_list += "<td>" + topic.Type + "</td>";
				topic_list += "</tr>";
			}
			$("#topic-list").html(topic_list);
			$("#form-create-item-topic-list").html(formCreateItemTopicList);
			$("#form-create-item").prop('action', "http://" + instance.host + ":" + instance.port + "/item");
			$("#form-create-item-submit").prop('disabled', false);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			alert("Error while loading information for " + instance.toString() + " : " + textStatus + " " + errorThrown);
		},
		dataType : "jsonp"
	});
}
function createItem() {
	var url = "http://"+currentInstance.host+":"+currentInstance.port+"/item";
	$('#form-create-item').ajaxForm({
		url 		: url,
		method 		: "POST",
		success 	: function(response) {
			$('#form-create-item-submit').prop('disabled', true);
			$('#form-create-item-alert').prop('color', "green");
			$('#form-create-item-alert').html("Created");
			setTimeout(function() {
				$('#form-create-item-alert').html("");
				$('#form-create-item-submit').prop('disabled', false);
			}, 1200);
		},
		error 		: function(jqXHR, textStatus, errorThrown) {
			$('#form-create-item-submit').prop('disabled', true);
			$('#form-create-item-alert').prop('color', "red");
			$('#form-create-item-alert').html("Erreur : " + errorThrown + " " + jqXHR.responseText);
			setTimeout(function() {
				$('#form-create-item-alert').html("");
				$('#form-create-item-submit').prop('disabled', false);
			}, 4000);
		}
	}).submit();
}
function addPropertyToNewItem() {
	$("#form-create-item-property-list").append("<tr><td><input name='property-name' style='width:100%' type='text'/></td><td><input name='property-value' style='width:100%' type='text'/></td><td style='text-align:center'><a href='#' class='button' onclick=\"$('#form-create-item-property-list').html('')\">X</a></td></tr>");
}
function clearItem() {
	$("#form-topic-item-id").val("");
	var html = "<table><thead><tr><td>Name</td><td>Value</td></tr></thead><tbody></tbody></table>";
	$("#form-topic-item-properties").html("");
	$("#form-topic-item-value").val("");
	$("#form-topic-item-alert").html("");
}
function popAnItem() {
	clearItem();
	var url = "http://" + currentInstance.host + ":" + currentInstance.port
			+ "/topic/" + currentTopic + "/pop";
	$.ajax({
		url : url,
		success : function(data, textStatus, jqXHR) {
			var ContentLength = jqXHR.getResponseHeader("Content-Length");
			var properties = jqXHR.getResponseHeader("Properties");
			properties = JSON.parse(properties);
			$("#form-topic-item-id").val(jqXHR.getResponseHeader("Id"));
			var html = "";
			for (var p = 0; p < properties.length; p++) {
				var property = properties[p];
				html += "<tr><td>" + property.name + "</td><td>"
						+ property.value + "</td></tr>";
			}
			$("#form-topic-item-properties").html(html);
			$("#form-topic-item-value").val(data);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			$("#form-topic-item-alert").html(errorThrown);
			setTimeout(function() {
				$("#form-topic-item-alert").html("");
			}, 1200);
		},
		dataType : "text"
	});
}
$(function() {
	$("#accordion").accordion({
		activate : function(event, ui) {
			loadInformation(toInstance(ui.newHeader.text()));
		}
	});
	$("#tabs").tabs();
	if (!store.isEmpty()) {
		reloadInstancesInStore();
	} else {
		$.ajax({
			url : "/info",
			success : function(data) {
				var instance = new Instance(data.Host, data.Port);
				instance.version = data.Version;
				addInstancePanel(instance, null);
				store.addInstance(instance);
				loadInformation(instance);
			},
			error : function(jqXHR, textStatus, errorThrown) {
				addInstancePanel(instance, "Unreachable " + errorThrown);
			},
			dataType : "jsonp"
		});
	}
});