package domain

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/events"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	// The events package's init registers the module, which Install looks up.
	jstest.Install(t, e, "events", events.Install)
	jstest.Install(t, e, "domain", Install)
	return e
}

func TestCreate(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		return String(d instanceof domain.Domain) + ':' +
			String(Array.isArray(d.members)) + ':' + String(domain.createDomain() instanceof domain.Domain);
	})()`), "true:true:true")
}

func TestRunEntersAndRestores(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		let inside = null;
		const result = d.run(() => {
			inside = domain.active;
			return 7;
		});
		return String(result) + ':' + String(inside === d) + ':' + String(domain.active);
	})()`), "7:true:null")
}

func TestRunRoutesThrownError(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		let message = 'none';
		d.on('error', (err) => { message = err.message; });
		d.run(() => { throw new Error('boom'); });
		return message;
	})()`), "boom")

	// Without an 'error' listener the error is not swallowed.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		try { d.run(() => { throw new Error('unhandled'); }); }
		catch (e) { return e.message; }
		return 'did not throw';
	})()`), "unhandled")
}

func TestAddRoutesEmitterErrors(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		const ee = new events.EventEmitter();
		let message = 'none';
		d.on('error', (err) => { message = err.message; });
		d.add(ee);
		const member = ee.domain === d;
		ee.emit('error', new Error('routed'));
		return String(member) + ':' + message;
	})()`), "true:routed")

	// Removing a member restores normal emitter behaviour.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		const ee = new events.EventEmitter();
		d.add(ee);
		d.remove(ee);
		return String(ee.domain) + ':' + String(d.members.length);
	})()`), "null:0")
}

func TestBindAndIntercept(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		let active = false;
		const bound = d.bind(() => { active = domain.active === d; });
		bound();
		return String(active) + ':' + String(domain.active);
	})()`), "true:null")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		let message = 'none';
		d.on('error', (err) => { message = err.message; });
		const intercepted = d.intercept((err) => { throw err; });
		intercepted(new Error('intercepted'));
		return message;
	})()`), "intercepted")

	// A clean first argument runs the callback inside the domain.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const d = domain.create();
		const intercepted = d.intercept((value) => String(domain.active === d) + ':' + value);
		return intercepted(null, 'ok');
	})()`), "true:ok")
}
