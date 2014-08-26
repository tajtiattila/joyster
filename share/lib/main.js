function GraphicUpdater(config) {
	var updaters = [];
	function getElem(n) {
		var elem;
		if (n in config) {
			var elem = config[n];
			if (typeof(elem) === "string") {
				elem = document.getElementById(elem);
			}
		}
		return elem;
	}
	function updater(n, fn) {
		var elem = getElem(n);
		if (elem) {
			updaters.push(fn(elem));
		}
	}
	function classUpdater(n, fn) {
		updater(n, function(elem){
			return function(data){
				fn(elem, data);
			}
		});
	}
	function stickCanvasUpdater(n, fn) {
		var elem = getElem(n);
		if (elem) {
			var cs = new CanvasStick({
				canvas: elem,
				trailMillis: config.trailMillis || 1000,
				trailStrokeStyle: config.trailStrokeStyle || "#aaa",
				crossStrokeStyle: config.crossStrokeStyle || "#000"
			});
			updaters.push(function(data){
				fn(cs, data);
			});
		}
	}
	var fd = config.fixDigits || 4;
	function valueDisplay(n, fn) {
		updater(n, function(elem) {
			return function(data) {
				var val = fn(data);
				var s = val.toFixed(fd);
				if (val > 0) {
					s = "+" + s;
				} else if (val == 0) {
					s = " " + s;
				}
				elem.innerHTML = s;
			}
		});
	}
	classUpdater("rollToYawDiv", function(elem, data){
		setClass(elem, "enabled", data.RollToYaw);
	});
	classUpdater("triggerYawDiv", function(elem, data){
		setClass(elem, "enabled", data.TriggerYaw);
	});
	classUpdater("headLookDiv", function(elem, data){
		setClass(elem, "enabled", data.HeadLook);
	});
	stickCanvasUpdater("inputLStickCanvas", function(cs, data){
		cs.add(data.I.LX, data.I.LY);
	});
	stickCanvasUpdater("inputRStickCanvas", function(cs, data){
		cs.add(data.I.RX, data.I.RY);
	});
	stickCanvasUpdater("outputXYCanvas", function(cs, data){
		cs.add(data.O.X, data.O.Y);
	});
	stickCanvasUpdater("outputZYCanvas", function(cs, data){
		cs.add(data.O.Z, data.O.Y);
	});
	stickCanvasUpdater("outputRXRYCanvas", function(cs, data){
		cs.add(data.O.RX, data.O.RY);
	});
	stickCanvasUpdater("outputUVCanvas", function(cs, data){
		cs.add(data.O.U, data.O.V);
	});
	valueDisplay("inputLX",  function(data) { return data.I.LX; });
	valueDisplay("inputLY",  function(data) { return data.I.LY; });
	valueDisplay("inputRX",  function(data) { return data.I.RX; });
	valueDisplay("inputRY",  function(data) { return data.I.RY; });
	valueDisplay("outputX",  function(data) { return data.O.X;  });
	valueDisplay("outputY",  function(data) { return data.O.Y;  });
	valueDisplay("outputZ",  function(data) { return data.O.Z;  });
	valueDisplay("outputRX", function(data) { return data.O.RX; });
	valueDisplay("outputRY", function(data) { return data.O.RY; });
	valueDisplay("outputRZ", function(data) { return data.O.RZ; });
	valueDisplay("outputU",  function(data) { return data.O.U;  });
	valueDisplay("outputV",  function(data) { return data.O.V;  });
	this.update = function(data) {
		for (var i = 0; i < updaters.length; i++) {
			updaters[i](data);
		}
	};
}
function CanvasStick(config) {
	var c = config.canvas;
	if (!c.getContext) {
		throw Error("bad id/no canvas support");
	}
	this.canvas = c;
	this.ctx = c.getContext("2d");
	this.ctx.scale(c.width/2.0, c.height/2.0);
	this.ctx.translate(1.0, 1.0);
	this.ctx.lineWidth = 2.0/c.width;
	this.trailMillis = config.trailMillis;
	this.trailStrokeStyle = config.trailStrokeStyle;
	this.crossStrokeStyle = config.crossStrokeStyle;
	this.crossdim = 0.1;
	this.trail = [];
	this.drawing = false;
	this.px = 0;
	this.py = 0;
	var csobj = this;
	window.setInterval(function() {
		var now = new Date().valueOf(); // milliseconds
		var old = now - csobj.trailMillis;
		while (csobj.trail.length && csobj.trail[0].timestamp < old) {
			csobj.trail.shift();
		}
		if (csobj.drawing)
			return;
		csobj.drawing = true;
		try {
			var ctx = csobj.ctx;
			ctx.clearRect(-1, -1, 2, 2);
			ctx.save();
			ctx.strokeStyle = csobj.trailStrokeStyle;
			ctx.beginPath();
			if (csobj.trail.length > 1) {
				ctx.moveTo(csobj.trail[0].x, csobj.trail[0].y);
				for (var i = 1; i < csobj.trail.length; i++) {
					ctx.lineTo(csobj.trail[i].x, csobj.trail[i].y);
				}
				ctx.stroke();
			}
			ctx.strokeStyle = csobj.crossStrokeStyle;
			ctx.beginPath();
			var s = csobj.crossdim;
			ctx.moveTo(csobj.px-s,csobj.py);
			ctx.lineTo(csobj.px+s,csobj.py);
			ctx.moveTo(csobj.px,csobj.py-s);
			ctx.lineTo(csobj.px,csobj.py+s);
			ctx.stroke();
			ctx.restore();
		}
		finally {
			csobj.drawing = false;
		}
	}, 50)
	this.add = function(x, y) {
		var now = new Date().valueOf(); // milliseconds
		this.px = x;
		this.py = -y;
		this.trail.push({x:this.px, y:this.py, timestamp:now});
	}
}
function websocket(config) {
	var elsock = config.elementSocketInfo;
	var reconnectDelay = 1000;
	if ('reconnectDelay' in config) {
		reconnectDelay = config.reconnectDelay;
	}
	var errorReportDelay = 5000;
	if ('errorReportDelay' in config) {
		errorReportDelay = config.errorReportDelay;
	}
	var connected = false;
	var conn;
	var gu = new GraphicUpdater(config);
	function connect() {
		conn = new WebSocket(config.url);
		conn.onopen = function(evt) {
			console.log('websocket connected');
			connected = true;
			elsock.innerHTML = '<p>Active</p>';
		}
		conn.onclose = function(evt) {
			console.log('websocket disconnected');
			connected = false;
			setTimeout(function(){
				connect();
			}, reconnectDelay);
			setTimeout(function(){
				if (!connected) {
					elsock.innerHTML = '<p>Inactive</p>';
				}
			}, errorReportDelay);
		}
		conn.onmessage = function(evt) {
			console.log('websocket updated');
			gu.update(JSON.parse(evt.data));
		}
	}
	connect();
}
function setClass(el, cls, flag) {
	if (flag) {
		addClass(el, cls);
	} else {
		removeClass(el, cls);
	}
}
function addClass(el, cls) {
	if (el.className == "") {
		el.className = cls;
		return
	}
	var v0 = el.className.split(" ");
	for (var i = 0; i < v0.length; i++) {
		if (v0[i] == cls)
			return;
	}
	v0.push(cls);
	el.className = v0.join(" ");
}
function removeClass(el, cls) {
	if (el.className == "") {
		return;
	}
	if (el.className == cls) {
		el.className = "";
		return;
	}
	var v0 = el.className.split(" ");
	var v1 = [];
	for (var i = 0; i < v0.length; i++) {
		if (v0[i] != "" && v0[i] != cls)
			v1.push(v0[i])
	}
	el.className = v1.join(" ");
}
