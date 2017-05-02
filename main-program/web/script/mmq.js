/**
 * Magic MQ admin script page.
 */
/**
 * Object Instance
 */
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
/**
 * Object Store
 */
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
	getInstance : function(aName) {
		var instances = this.getInstances();
		for (var i = 0; i < instances.length; i++) {
			if ((instances[i].host+":"+instances[i].port) == aName) {
				return instances[i];
			}
		}
		return null;
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
//
// Display fonctions
//

function addInstancePanel(instance, error) {
	var html = '<div><p>';
	if (error != null) {
		html += 'Unable to connect ' + error + '<br/>';
		// html += '<button
		// onclick="loadInstance(\''+instance.toString()+'\',false)"></button><br/>';
	} else {
		html += '<b>Version</b> : ' + instance.version + '<br/>';
	}
	html += '<b>Groups</b> : ';
	if (typeof instance.groups != "undefined") {
		for (var i = 0; i < instance.groups.length; i++){
			if (i != 0){
				html += ", ";
			}
			html += instance.groups[i];
		}
	}
	html += '<br/>'
	html += '<a href="#" class="button">Remove instance</a>';
	$('#accordion').append('<h3>' + instance.toString() + '</h3>' + html + '</p></div>').accordion("refresh");
}
function shutdown() {
	if (!confirm("Voulez-vous vraiment arrêter ce serveur ?")){
		return;
	}
	var url = "http://" + currentInstance.host + ":" + currentInstance.port + "/shutdown";
	$.ajax({
		url : url,
		success : function(data) {
			/**
			 * var instance = new Instance(data.Host, data.Port);
			 * instance.version = data.Version; instance.groups = data.Groups;
			 * addInstancePanel(instance, null); if (addToList) {
			 * store.addInstance(instance); }
			 */
		},
		error : function(jqXHR, textStatus, errorThrown) {
			// addInstancePanel(instance, "Unreachable " + errorThrown);
		},
		dataType : "json"
	});
}
function loadInstance(aInstance, addToList) {
	var url = "http://" + aInstance.host + ":" + aInstance.port + "/info";
	$.ajax({
		url : url,
		success : function(data) {
			var instance = new Instance(data.Host, data.Port);
			instance.version = data.Version;
			instance.groups = data.Groups;
			addInstancePanel(instance, null);
			if (addToList) {
				store.addInstance(instance);
			}
		},
		error : function(jqXHR, textStatus, errorThrown) {
			addInstancePanel(aInstance, "Unreachable " + errorThrown);
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
			loadInstanceInformation(instance);
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
	$("#form-topic-button-pop").prop('disabled', true);
	$("#form-topic-button-list").prop('disabled', true);
	currentTopic = null;
	$.ajax({
		url : "http://" + currentInstance.host + ":" + currentInstance.port + "/topic/" + aTopicName,
		success : function(data) {
			currentTopic = data.Name;
			$("#tabs").tabs("option", "active", 2);
			$("#form-topic-title").html(data.Name);
			$("#form-topic-button-pop").prop('disabled', false);
			$("#form-topic-button-list").prop('disabled', false);
			$("#form-topic-type").val(data.Type);
			$("#form-topic-item-id").html("");
			$("#form-topic-item-properties").html("");
			$("#form-topic-item-value").val("");
			$("#form-topic-item-list").html("");
			$("#form-topic-rsslink").prop("href","/topic/"+aTopicName+"/rss");
			var html = "";
			for (var i = 0; i < data.Parameters.length; i++) {
				var property = data.Parameters[i];
				html+="<tr><td>"+property.Name+"</td><td>"+property.Value+"</td></tr>";
			}
			$("#form-topic-property-list").html(html);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			alert("Error " + textStatus + " " + errorThrown);
		},
		dataType : "jsonp"
	})
}
function loadLogs(){
	var url = "http://"+currentInstance.host+":"+currentInstance.port+"/log";
	$('#instance-logs').prop('src',url);
}
function loadInstanceInformation(aInstance) {
	$("#form-topic-title").html("");
	$("#form-topic-button-pop").prop('disabled', true);
	$("#form-topic-button-list").prop('disabled', true);
	$("#form-create-item-submit").prop('disabled', true);
	$("#form-config-title").html(aInstance.host+":"+aInstance.port);
	currentInstance = aInstance;
	loadLogs();
	$.ajax({
		url : "http://" + aInstance.host + ":" + aInstance.port + "/info",
		success : function(data) {
			$("#form-config-version").val(data.Version);
		},
		error : function(jqXHR, textStatus, errorThrown) {
			addInstancePanel(aInstance, "Unreachable " + errorThrown);
		},
		dataType : "json"
	});
	$.ajax({
		url : "http://" + aInstance.host + ":" + aInstance.port + "/topic",
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
			$("#form-create-item").prop('action', "http://" + aInstance.host + ":" + aInstance.port + "/item");
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
	var radios = $('input[type=radio][name=form-create-item-content-type]');
	var mode = "text";
	for (var i = 0; i < radios.length ; i++) {
		if (radios[i].checked) {
			mode = radios[i].value;
		}
	}
	var query = {
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
		};
	var enctype;
	if (mode == "file") {
		var value = $("#form-create-item-as-file").val();
		if (value == ""){
			alert("File is missing");
			return;
		}
		$('#form-create-item').ajaxForm(query).submit();
	} else {
		var data = new Object();
		var inputs = $("#form-create-item * input" );
		for (var i = 0; i < inputs.length; i++){
			if (inputs[i].name == "value") {
				continue;
			}
			if (inputs[i].type == "checkbox") {
				if (!inputs[i].checked) {
					continue;
				}
			}
			if (typeof data[inputs[i].name] == "undefined") {
				data[inputs[i].name] = new Array(inputs[i].value);
			} else {
				data[inputs[i].name].push(inputs[i].value);
			}
		}
		data.value = $("#form-create-item-as-text").val();
		query.data = data;
		query.traditional = true; // necessary for values list sending
		$.ajax(query);
	}
	
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
function listItems() {
	var url = "http://" + currentInstance.host + ":" + currentInstance.port + "/topic/" + currentTopic + "/list";
	$("#form-topic-item-list").html("");
	$.ajax({
		url : url,
		success : function(items, textStatus, jqXHR) {
			if (items != null) { 
				html = "";
				for (var i = 0; i < items.length; i++) {
					item = items[i];
					var properties = "";
					if (item.Properties != null) {
						for (var p = 0; p < item.Properties.length; p++){
							if (p > 0){
								properties += " ; ";
							}
							properties += "<b>"+item.Properties[p].Name+"</b> : "+item.Properties[p].Value;
						}
					}
					html += "<tr><td>"+item.ID+"</td><td>"+Math.round(parseInt(item.Age)/1000000)+" ms</td><td>"+properties+"</td></tr>";
				}
				$("#form-topic-item-list").html(html);
			}
		},
		error : function(jqXHR, textStatus, errorThrown) {
			$("#form-topic-item-alert").html(errorThrown);
			setTimeout(function() {
				$("#form-topic-item-alert").html("");
			}, 1200);
		},
		dataType : "json"
	});
}
function popAnItem() {
	clearItem();
	var url = "http://" + currentInstance.host + ":" + currentInstance.port	+ "/topic/" + currentTopic + "/pop";
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
			if ($("#form-topic-item-list").html() != ""){
				listItems()
			}
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
			var instance = store.getInstance(ui.newHeader.text());
			loadInstanceInformation(instance);
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
				instance.groups = data.Groups;
				addInstancePanel(instance, null);
				store.addInstance(instance);
				loadInstanceInformation(instance);
			},
			error : function(jqXHR, textStatus, errorThrown) {
				addInstancePanel(instance, "Unreachable " + errorThrown);
			},
			dataType : "jsonp"
		});
	}
});