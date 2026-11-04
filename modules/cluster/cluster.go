// Package cluster implements the primary/worker fork surface on top of
// child_process.fork(). Workers are ordinary forked processes; the marker in
// their environment makes the same cluster module identify them as workers.
package cluster

import (
	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/child_process"
	_ "godejs/modules/events"
)

func init() { env.Register("cluster", Install) }

func Install(e *env.Env) js.Object {
	childProcess := e.Exports("child_process")
	events := e.Exports("events")
	ee := events.Get("EventEmitter")
	if !ee.Valid() || !ee.IsFunction() {
		panic("godejs cluster: events.EventEmitter is unavailable")
	}
	api, err := e.JS.Eval(prelude, "cluster.js")
	if err != nil {
		panic("godejs cluster: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, childProcess.V(), e.Global().Get("process"), ee)
	if err != nil {
		panic("godejs cluster: " + err.Error())
	}
	m := built.Object()
	m.SetBool("isMaster", !m.Get("isWorker").Bool())
	return m
}

func Implemented() []string {
	return []string{
		"cluster.Worker",
		"cluster.Worker#disconnect",
		"cluster.Worker#kill",
		"cluster.Worker#send",
		"cluster.disconnect",
		"cluster.fork",
		"cluster.isMaster",
		"cluster.isPrimary",
		"cluster.isWorker",
		"cluster.schedulingPolicy",
		"cluster.settings",
		"cluster.setupMaster",
		"cluster.setupPrimary",
		"cluster.workers",
	}
}

const prelude = `(function (childProcess, process, EventEmitter) {
  'use strict';

  var primary = !(process && process.env && process.env.GODE_CLUSTER_WORKER_ID);

  function settingsObject() {
    return {
      args: process && process.argv ? process.argv.slice(2) : [],
      exec: process && process.argv ? process.argv[1] : undefined,
      execArgv: process && process.execArgv ? process.execArgv.slice() : [],
      silent: false,
      stdio: 'pipe',
      uid: undefined,
      gid: undefined,
      inspectPort: undefined,
      windowsHide: false,
      cwd: undefined,
      serialization: 'json'
    };
  }

  var settings = settingsObject();
  var workers = Object.create(null);
  var nextWorkerID = 1;

  class Worker extends EventEmitter {
    constructor(id, child) {
      super();
      this.id = id;
      this.process = child;
      this.state = 'online';
      this.exitedAfterDisconnect = false;
      this._disconnecting = false;
      var self = this;
      child.on('message', function (message, handle) {
        self.emit('message', message, handle);
      });
      child.on('disconnect', function () {
        self.state = self._disconnecting ? 'disconnected' : self.state;
        self.emit('disconnect');
      });
      child.on('exit', function (code, signal) {
        self.state = 'dead';
        self.exitedAfterDisconnect = self._disconnecting;
        self.emit('exit', code, signal);
        delete workers[id];
      });
      child.on('error', function (err) { self.emit('error', err); });
    }
    send(message) {
      this.process.send.apply(this.process, arguments);
      return this;
    }
    disconnect() {
      this._disconnecting = true;
      this.state = 'disconnecting';
      this.process.disconnect();
      return this;
    }
    kill(signal) {
      this.process.kill(signal);
      return this;
    }
    get isConnected() { return !!this.process.connected; }
  }

  function setupPrimary(options) {
    if (options !== undefined && (!options || typeof options !== 'object')) {
      var err = new TypeError('The "settings" argument must be of type object');
      err.code = 'ERR_INVALID_ARG_TYPE';
      throw err;
    }
    var next = settingsObject();
    for (var sk in settings) {
      if (Object.prototype.hasOwnProperty.call(settings, sk)) next[sk] = settings[sk];
    }
    if (options) {
      for (var key in options) {
        if (Object.prototype.hasOwnProperty.call(options, key)) next[key] = options[key];
      }
    }
    if (next.execArgv && !Array.isArray(next.execArgv)) next.execArgv = [];
    if (next.args && !Array.isArray(next.args)) next.args = [];
    settings = next;
    cluster.settings = settings;
    if (cluster.schedulingPolicy !== 1 && cluster.schedulingPolicy !== 2) {
      var policyErr = new TypeError('Bad cluster.schedulingPolicy: ' + cluster.schedulingPolicy);
      policyErr.code = 'ERR_INVALID_ARG_VALUE';
      throw policyErr;
    }
    var emitSetup = function () { cluster.emit('setup', settings); };
    if (process && typeof process.nextTick === 'function') process.nextTick(emitSetup);
    else Promise.resolve().then(emitSetup);
    return settings;
  }

  function fork(env) {
    if (!primary) {
      var err = new Error('Cannot fork a worker from a worker');
      err.code = 'ERR_WORKER_INIT_FAILED';
      throw err;
    }
    if (!settings.exec) {
      var noExec = new Error('The cluster exec path is not set');
      noExec.code = 'ERR_CLUSTER_NO_EXEC';
      throw noExec;
    }
    var id = nextWorkerID++;
    var childEnv = {};
    if (process && process.env) {
      for (var key in process.env) {
        if (Object.prototype.hasOwnProperty.call(process.env, key)) childEnv[key] = process.env[key];
      }
    }
    if (env && typeof env === 'object') {
      for (var ekey in env) {
        if (Object.prototype.hasOwnProperty.call(env, ekey)) childEnv[ekey] = env[ekey];
      }
    }
    childEnv.GODE_CLUSTER_WORKER_ID = String(id);
    childEnv.NODE_UNIQUE_ID = String(id);
    var opts = {
      env: childEnv,
      execArgv: settings.execArgv,
      silent: settings.silent,
      serialization: settings.serialization
    };
    if (settings.cwd !== undefined) opts.cwd = settings.cwd;
    if (settings.uid !== undefined) opts.uid = settings.uid;
    if (settings.gid !== undefined) opts.gid = settings.gid;
    if (settings.windowsHide !== undefined) opts.windowsHide = settings.windowsHide;
    if (settings.stdio !== undefined) opts.stdio = settings.stdio;
    var child = childProcess.fork(settings.exec, settings.args || [], opts);
    var worker = new Worker(id, child);
    workers[id] = worker;
    cluster.emit('fork', worker);
    return worker;
  }

  function disconnect(callback) {
    var list = Object.keys(workers);
    if (!list.length) {
      if (typeof callback === 'function') process.nextTick(callback);
      return;
    }
    var pending = list.length;
    function done() {
      if (--pending === 0 && typeof callback === 'function') callback();
    }
    for (var i = 0; i < list.length; i++) {
      var worker = workers[list[i]];
      if (!worker) { done(); continue; }
      worker.once('exit', done);
      worker.disconnect();
    }
  }

  var SCHED_NONE = 1;
  var SCHED_RR = 2;
  var policy = process && process.env ? process.env.NODE_CLUSTER_SCHED_POLICY : undefined;
  if (policy === 'none') policy = SCHED_NONE;
  else if (policy === 'rr') policy = SCHED_RR;
  else if (process && process.platform === 'win32') policy = SCHED_NONE;
  else policy = SCHED_RR;

  var cluster = new EventEmitter();
  cluster.isPrimary = primary;
  cluster.isMaster = primary;
  cluster.isWorker = !primary;
  cluster.SCHED_NONE = SCHED_NONE;
  cluster.SCHED_RR = SCHED_RR;
  cluster.schedulingPolicy = policy;
  cluster.settings = settings;
  cluster.workers = workers;
  cluster.Worker = Worker;
  cluster.fork = fork;
  cluster.disconnect = disconnect;
  cluster.setupPrimary = setupPrimary;
  cluster.setupMaster = setupPrimary;
  if (!primary) {
    var wid = Number(process.env.GODE_CLUSTER_WORKER_ID || process.env.NODE_UNIQUE_ID || 0);
    cluster.worker = { id: wid, process: process, state: 'online', send: function () { return this; }, disconnect: function () { return this; }, kill: function () { return this; } };
  } else {
    cluster.worker = null;
  }
  if (!primary) {
    cluster.fork = function () {
      var err = new Error('Cannot fork a worker from a worker');
      err.code = 'ERR_WORKER_INIT_FAILED';
      throw err;
    };
  }
  return cluster;
})`
