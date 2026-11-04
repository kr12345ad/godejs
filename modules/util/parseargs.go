package util

import (
	"fmt"

	"godejs/internal/env"
	"godejs/internal/js"
)

// installParseArgs adds util.parseArgs. The whole option grammar lives in
// JavaScript so the host only has to hand over the exports object.
func installParseArgs(e *env.Env, exports js.Object) {
	factory, err := e.JS.Eval(parseArgsPrelude, "[gode:util:parseArgs]")
	if err != nil {
		fmt.Fprintf(e.Err(), "util parseArgs prelude: %v\n", err)
		return
	}
	if _, err := factory.Object().Call(js.Object{}, exports.V()); err != nil {
		fmt.Fprintf(e.Err(), "util parseArgs prelude: %v\n", err)
	}
}

const parseArgsPrelude = `(function (exports) {
  'use strict';

  function describe(value) {
    if (value === null) return 'null';
    var t = typeof value;
    if (t === 'undefined') return 'undefined';
    if (t === 'string') return "type string ('" + value + "')";
    if (t === 'number') return 'type number (' + value + ')';
    if (t === 'boolean') return 'type boolean (' + value + ')';
    if (t === 'function') return 'function ' + (value.name || '(anonymous)');
    if (Array.isArray(value)) return 'an instance of Array';
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }

  function invalidType(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function parseError(code, message) {
    var err = new TypeError(message);
    err.code = code;
    return err;
  }

  function parseArgs(config) {
    if (config === undefined || config === null) config = {};
    if (typeof config !== 'object' || Array.isArray(config)) {
      throw invalidType('config', 'of type object', config);
    }
    var args = config.args === undefined ? process.argv.slice(2) : config.args;
    if (!Array.isArray(args)) throw invalidType('config.args', 'an instance of Array', args);
    for (var ai = 0; ai < args.length; ai++) {
      if (typeof args[ai] !== 'string') {
        throw invalidType('config.args[' + ai + ']', 'of type string', args[ai]);
      }
    }
    var strict = config.strict === undefined ? true : config.strict;
    if (typeof strict !== 'boolean') throw invalidType('config.strict', 'of type boolean', strict);
    var allowPositionals = config.allowPositionals === undefined ? false : config.allowPositionals;
    if (typeof allowPositionals !== 'boolean') {
      throw invalidType('config.allowPositionals', 'of type boolean', allowPositionals);
    }
    var returnTokens = config.tokens === undefined ? false : config.tokens;
    if (typeof returnTokens !== 'boolean') throw invalidType('config.tokens', 'of type boolean', returnTokens);

    var options = config.options === undefined ? {} : config.options;
    if (options === null || typeof options !== 'object' || Array.isArray(options)) {
      throw invalidType('config.options', 'of type object', options);
    }

    var specs = Object.create(null);
    var shorts = Object.create(null);
    var names = Object.keys(options);
    for (var i = 0; i < names.length; i++) {
      var name = names[i];
      var raw = options[name] === undefined ? {} : options[name];
      if (raw === null || typeof raw !== 'object' || Array.isArray(raw)) {
        throw invalidType('config.options.' + name, 'of type object', raw);
      }
      var type = raw.type === undefined ? 'boolean' : raw.type;
      if (type !== 'string' && type !== 'boolean') {
        var typeErr = new TypeError("The argument 'options." + name +
          ".type' must be one of: 'string', 'boolean'. Received '" + type + "'");
        typeErr.code = 'ERR_INVALID_ARG_VALUE';
        throw typeErr;
      }
      if (raw.multiple !== undefined && typeof raw.multiple !== 'boolean') {
        throw invalidType('config.options.' + name + '.multiple', 'of type boolean', raw.multiple);
      }
      if (raw.default !== undefined && raw.multiple !== true) {
        if (type === 'boolean' && typeof raw.default !== 'boolean') {
          throw invalidType('config.options.' + name + '.default', 'of type boolean', raw.default);
        }
        if (type === 'string' && typeof raw.default !== 'string') {
          throw invalidType('config.options.' + name + '.default', 'of type string', raw.default);
        }
      }
      if (raw.short !== undefined) {
        if (typeof raw.short !== 'string' || raw.short.length !== 1) {
          var shortErr = new TypeError("The argument 'options." + name +
            ".short' must be a single character. Received '" + raw.short + "'");
          shortErr.code = 'ERR_INVALID_ARG_VALUE';
          throw shortErr;
        }
        shorts[raw.short] = name;
      }
      specs[name] = {
        name: name,
        type: type,
        multiple: raw.multiple === true,
        defaultValue: raw.default,
        hadDefault: raw.default !== undefined
      };
    }

    var values = {};
    for (var sn = 0; sn < names.length; sn++) {
      var base = specs[names[sn]];
      if (base.multiple) values[base.name] = base.hadDefault ? [].concat(base.defaultValue) : [];
      else if (base.hadDefault) values[base.name] = base.defaultValue;
    }

    var positionals = [];
    var tokens = [];

    function setOption(spec, value, rawName) {
      if (spec.multiple) values[spec.name].push(value);
      else values[spec.name] = value;
      if (returnTokens) tokens.push({ kind: 'option', name: spec.name, rawName: rawName, value: value });
    }

    var idx = 0;
    var onlyPositionals = false;
    while (idx < args.length) {
      var arg = args[idx];
      if (!onlyPositionals && arg === '--') {
        onlyPositionals = true;
        idx++;
        continue;
      }
      if (!onlyPositionals && arg.length > 2 && arg.slice(0, 2) === '--') {
        var body = arg.slice(2);
        var eq = body.indexOf('=');
        var rawName = eq >= 0 ? body.slice(0, eq) : body;
        var optName = rawName;
        var inlineValue = eq >= 0 ? body.slice(eq + 1) : undefined;
        var negated = false;
        if (optName.slice(0, 3) === 'no-') {
          var positive = optName.slice(3);
          if (specs[positive] && specs[positive].type === 'boolean') {
            optName = positive;
            negated = true;
          }
        }
        var spec = specs[optName];
        if (!spec) {
          if (strict) {
            throw parseError('ERR_PARSE_ARGS_UNKNOWN_OPTION', "Unknown option '--" + rawName + "'");
          }
          positionals.push(arg);
          if (returnTokens) tokens.push({ kind: 'positional', index: positionals.length - 1, value: arg });
          idx++;
          continue;
        }
        if (spec.type === 'boolean') {
          var boolValue = negated ? false : true;
          if (inlineValue !== undefined) {
            if (inlineValue === 'true') boolValue = true;
            else if (inlineValue === 'false') boolValue = false;
            else {
              throw parseError('ERR_PARSE_ARGS_INVALID_OPTION_VALUE',
                "Option '--" + rawName + '=' + inlineValue + "' does not take an argument");
            }
          }
          setOption(spec, boolValue, '--' + rawName);
          idx++;
          continue;
        }
        var value = inlineValue;
        if (value === undefined) {
          if (idx + 1 >= args.length) {
            throw parseError('ERR_PARSE_ARGS_INVALID_OPTION_VALUE',
              "Option '--" + rawName + " <value>' argument missing");
          }
          value = args[idx + 1];
          idx++;
        }
        setOption(spec, value, '--' + rawName);
        idx++;
        continue;
      }
      if (!onlyPositionals && arg.length > 1 && arg[0] === '-' && arg !== '-') {
        var group = arg.slice(1);
        var consumedNext = false;
        for (var gi = 0; gi < group.length; gi++) {
          var ch = group[gi];
          var shortName = shorts[ch];
          if (!shortName) {
            if (strict) throw parseError('ERR_PARSE_ARGS_UNKNOWN_OPTION', "Unknown option '-" + ch + "'");
            positionals.push(arg);
            if (returnTokens) tokens.push({ kind: 'positional', index: positionals.length - 1, value: arg });
            gi = group.length;
            continue;
          }
          var shortSpec = specs[shortName];
          if (shortSpec.type === 'boolean') {
            setOption(shortSpec, true, '-' + ch);
            continue;
          }
          var shortValue = group.slice(gi + 1);
          if (shortValue.slice(0, 1) === '=') shortValue = shortValue.slice(1);
          if (shortValue === '') {
            if (idx + 1 < args.length) {
              shortValue = args[idx + 1];
              consumedNext = true;
            } else {
              shortValue = undefined;
            }
          }
          if (shortValue === undefined) {
            throw parseError('ERR_PARSE_ARGS_INVALID_OPTION_VALUE',
              "Option '-" + ch + " <value>' argument missing");
          }
          setOption(shortSpec, shortValue, '-' + ch);
          gi = group.length;
        }
        idx += consumedNext ? 2 : 1;
        continue;
      }
      if (!allowPositionals && strict) {
        throw parseError('ERR_PARSE_ARGS_UNEXPECTED_POSITIONAL',
          "Unexpected argument '" + arg + "'. This command does not take positional arguments");
      }
      positionals.push(arg);
      if (returnTokens) tokens.push({ kind: 'positional', index: positionals.length - 1, value: arg });
      idx++;
    }

    var result = { values: values, positionals: positionals };
    if (returnTokens) result.tokens = tokens;
    return result;
  }

  exports.parseArgs = parseArgs;
})`
