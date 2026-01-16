import { __commonJS, __toESM } from "./chunk-BlwiYZwt.js";
import { APICallError, EmptyResponseBodyError, InvalidArgumentError, InvalidPromptError, InvalidResponseDataError, JSONParseError, TooManyEmbeddingValuesForCallError, TypeValidationError, UnsupportedFunctionalityError, z } from "./react-C-QzwZjq.js";
import { existsSync, promises } from "node:fs";
import { dirname, join, resolve } from "node:path";

//#region node_modules/destr/dist/index.mjs
const suspectProtoRx$1 = /"(?:_|\\u0{2}5[Ff]){2}(?:p|\\u0{2}70)(?:r|\\u0{2}72)(?:o|\\u0{2}6[Ff])(?:t|\\u0{2}74)(?:o|\\u0{2}6[Ff])(?:_|\\u0{2}5[Ff]){2}"\s*:/;
const suspectConstructorRx$1 = /"(?:c|\\u0063)(?:o|\\u006[Ff])(?:n|\\u006[Ee])(?:s|\\u0073)(?:t|\\u0074)(?:r|\\u0072)(?:u|\\u0075)(?:c|\\u0063)(?:t|\\u0074)(?:o|\\u006[Ff])(?:r|\\u0072)"\s*:/;
const JsonSigRx = /^\s*["[{]|^\s*-?\d{1,16}(\.\d{1,17})?([Ee][+-]?\d+)?\s*$/;
function jsonParseTransform(key, value) {
	if (key === "__proto__" || key === "constructor" && value && typeof value === "object" && "prototype" in value) {
		warnKeyDropped(key);
		return;
	}
	return value;
}
function warnKeyDropped(key) {
	console.warn(`[destr] Dropping "${key}" key to prevent prototype pollution.`);
}
function destr(value, options = {}) {
	if (typeof value !== "string") return value;
	const _value = value.trim();
	if (value[0] === "\"" && value.endsWith("\"") && !value.includes("\\")) return _value.slice(1, -1);
	if (_value.length <= 9) {
		const _lval = _value.toLowerCase();
		if (_lval === "true") return true;
		if (_lval === "false") return false;
		if (_lval === "undefined") return void 0;
		if (_lval === "null") return null;
		if (_lval === "nan") return Number.NaN;
		if (_lval === "infinity") return Number.POSITIVE_INFINITY;
		if (_lval === "-infinity") return Number.NEGATIVE_INFINITY;
	}
	if (!JsonSigRx.test(value)) {
		if (options.strict) throw new SyntaxError("[destr] Invalid JSON");
		return value;
	}
	try {
		if (suspectProtoRx$1.test(value) || suspectConstructorRx$1.test(value)) {
			if (options.strict) throw new Error("[destr] Possible prototype pollution");
			return JSON.parse(value, jsonParseTransform);
		}
		return JSON.parse(value);
	} catch (error) {
		if (options.strict) throw error;
		return value;
	}
}

//#endregion
//#region node_modules/unstorage/dist/shared/unstorage.mNKHTF5Y.mjs
function wrapToPromise(value) {
	if (!value || typeof value.then !== "function") return Promise.resolve(value);
	return value;
}
function asyncCall(function_, ...arguments_) {
	try {
		return wrapToPromise(function_(...arguments_));
	} catch (error) {
		return Promise.reject(error);
	}
}
function isPrimitive(value) {
	const type = typeof value;
	return value === null || type !== "object" && type !== "function";
}
function isPureObject(value) {
	const proto = Object.getPrototypeOf(value);
	return !proto || proto.isPrototypeOf(Object);
}
function stringify(value) {
	if (isPrimitive(value)) return String(value);
	if (isPureObject(value) || Array.isArray(value)) return JSON.stringify(value);
	if (typeof value.toJSON === "function") return stringify(value.toJSON());
	throw new Error("[unstorage] Cannot stringify value!");
}
const BASE64_PREFIX = "base64:";
function serializeRaw(value) {
	if (typeof value === "string") return value;
	return BASE64_PREFIX + base64Encode(value);
}
function deserializeRaw(value) {
	if (typeof value !== "string") return value;
	if (!value.startsWith(BASE64_PREFIX)) return value;
	return base64Decode(value.slice(BASE64_PREFIX.length));
}
function base64Decode(input) {
	if (globalThis.Buffer) return Buffer.from(input, "base64");
	return Uint8Array.from(globalThis.atob(input), (c) => c.codePointAt(0));
}
function base64Encode(input) {
	if (globalThis.Buffer) return Buffer.from(input).toString("base64");
	return globalThis.btoa(String.fromCodePoint(...input));
}
function normalizeKey(key) {
	if (!key) return "";
	return key.split("?")[0]?.replace(/[/\\]/g, ":").replace(/:+/g, ":").replace(/^:|:$/g, "") || "";
}
function joinKeys(...keys) {
	return normalizeKey(keys.join(":"));
}
function normalizeBaseKey(base) {
	base = normalizeKey(base);
	return base ? base + ":" : "";
}
function filterKeyByDepth(key, depth) {
	if (depth === void 0) return true;
	let substrCount = 0;
	let index = key.indexOf(":");
	while (index > -1) {
		substrCount++;
		index = key.indexOf(":", index + 1);
	}
	return substrCount <= depth;
}
function filterKeyByBase(key, base) {
	if (base) return key.startsWith(base) && key[key.length - 1] !== "$";
	return key[key.length - 1] !== "$";
}

//#endregion
//#region node_modules/unstorage/dist/index.mjs
function defineDriver$1(factory) {
	return factory;
}
const DRIVER_NAME$1 = "memory";
const memory = defineDriver$1(() => {
	const data = /* @__PURE__ */ new Map();
	return {
		name: DRIVER_NAME$1,
		getInstance: () => data,
		hasItem(key) {
			return data.has(key);
		},
		getItem(key) {
			return data.get(key) ?? null;
		},
		getItemRaw(key) {
			return data.get(key) ?? null;
		},
		setItem(key, value) {
			data.set(key, value);
		},
		setItemRaw(key, value) {
			data.set(key, value);
		},
		removeItem(key) {
			data.delete(key);
		},
		getKeys() {
			return [...data.keys()];
		},
		clear() {
			data.clear();
		},
		dispose() {
			data.clear();
		}
	};
});
function createStorage(options = {}) {
	const context = {
		mounts: { "": options.driver || memory() },
		mountpoints: [""],
		watching: false,
		watchListeners: [],
		unwatch: {}
	};
	const getMount = (key) => {
		for (const base of context.mountpoints) if (key.startsWith(base)) return {
			base,
			relativeKey: key.slice(base.length),
			driver: context.mounts[base]
		};
		return {
			base: "",
			relativeKey: key,
			driver: context.mounts[""]
		};
	};
	const getMounts = (base, includeParent) => {
		return context.mountpoints.filter((mountpoint) => mountpoint.startsWith(base) || includeParent && base.startsWith(mountpoint)).map((mountpoint) => ({
			relativeBase: base.length > mountpoint.length ? base.slice(mountpoint.length) : void 0,
			mountpoint,
			driver: context.mounts[mountpoint]
		}));
	};
	const onChange = (event, key) => {
		if (!context.watching) return;
		key = normalizeKey(key);
		for (const listener of context.watchListeners) listener(event, key);
	};
	const startWatch = async () => {
		if (context.watching) return;
		context.watching = true;
		for (const mountpoint in context.mounts) context.unwatch[mountpoint] = await watch(context.mounts[mountpoint], onChange, mountpoint);
	};
	const stopWatch = async () => {
		if (!context.watching) return;
		for (const mountpoint in context.unwatch) await context.unwatch[mountpoint]();
		context.unwatch = {};
		context.watching = false;
	};
	const runBatch = (items, commonOptions, cb) => {
		const batches = /* @__PURE__ */ new Map();
		const getBatch = (mount) => {
			let batch = batches.get(mount.base);
			if (!batch) {
				batch = {
					driver: mount.driver,
					base: mount.base,
					items: []
				};
				batches.set(mount.base, batch);
			}
			return batch;
		};
		for (const item of items) {
			const isStringItem = typeof item === "string";
			const key = normalizeKey(isStringItem ? item : item.key);
			const value = isStringItem ? void 0 : item.value;
			const options2 = isStringItem || !item.options ? commonOptions : {
				...commonOptions,
				...item.options
			};
			const mount = getMount(key);
			getBatch(mount).items.push({
				key,
				value,
				relativeKey: mount.relativeKey,
				options: options2
			});
		}
		return Promise.all([...batches.values()].map((batch) => cb(batch))).then((r) => r.flat());
	};
	const storage = {
		hasItem(key, opts = {}) {
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			return asyncCall(driver.hasItem, relativeKey, opts);
		},
		getItem(key, opts = {}) {
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			return asyncCall(driver.getItem, relativeKey, opts).then((value) => destr(value));
		},
		getItems(items, commonOptions = {}) {
			return runBatch(items, commonOptions, (batch) => {
				if (batch.driver.getItems) return asyncCall(batch.driver.getItems, batch.items.map((item) => ({
					key: item.relativeKey,
					options: item.options
				})), commonOptions).then((r) => r.map((item) => ({
					key: joinKeys(batch.base, item.key),
					value: destr(item.value)
				})));
				return Promise.all(batch.items.map((item) => {
					return asyncCall(batch.driver.getItem, item.relativeKey, item.options).then((value) => ({
						key: item.key,
						value: destr(value)
					}));
				}));
			});
		},
		getItemRaw(key, opts = {}) {
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			if (driver.getItemRaw) return asyncCall(driver.getItemRaw, relativeKey, opts);
			return asyncCall(driver.getItem, relativeKey, opts).then((value) => deserializeRaw(value));
		},
		async setItem(key, value, opts = {}) {
			if (value === void 0) return storage.removeItem(key);
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			if (!driver.setItem) return;
			await asyncCall(driver.setItem, relativeKey, stringify(value), opts);
			if (!driver.watch) onChange("update", key);
		},
		async setItems(items, commonOptions) {
			await runBatch(items, commonOptions, async (batch) => {
				if (batch.driver.setItems) return asyncCall(batch.driver.setItems, batch.items.map((item) => ({
					key: item.relativeKey,
					value: stringify(item.value),
					options: item.options
				})), commonOptions);
				if (!batch.driver.setItem) return;
				await Promise.all(batch.items.map((item) => {
					return asyncCall(batch.driver.setItem, item.relativeKey, stringify(item.value), item.options);
				}));
			});
		},
		async setItemRaw(key, value, opts = {}) {
			if (value === void 0) return storage.removeItem(key, opts);
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			if (driver.setItemRaw) await asyncCall(driver.setItemRaw, relativeKey, value, opts);
			else if (driver.setItem) await asyncCall(driver.setItem, relativeKey, serializeRaw(value), opts);
			else return;
			if (!driver.watch) onChange("update", key);
		},
		async removeItem(key, opts = {}) {
			if (typeof opts === "boolean") opts = { removeMeta: opts };
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			if (!driver.removeItem) return;
			await asyncCall(driver.removeItem, relativeKey, opts);
			if (opts.removeMeta || opts.removeMata) await asyncCall(driver.removeItem, relativeKey + "$", opts);
			if (!driver.watch) onChange("remove", key);
		},
		async getMeta(key, opts = {}) {
			if (typeof opts === "boolean") opts = { nativeOnly: opts };
			key = normalizeKey(key);
			const { relativeKey, driver } = getMount(key);
			const meta = /* @__PURE__ */ Object.create(null);
			if (driver.getMeta) Object.assign(meta, await asyncCall(driver.getMeta, relativeKey, opts));
			if (!opts.nativeOnly) {
				const value = await asyncCall(driver.getItem, relativeKey + "$", opts).then((value_) => destr(value_));
				if (value && typeof value === "object") {
					if (typeof value.atime === "string") value.atime = new Date(value.atime);
					if (typeof value.mtime === "string") value.mtime = new Date(value.mtime);
					Object.assign(meta, value);
				}
			}
			return meta;
		},
		setMeta(key, value, opts = {}) {
			return this.setItem(key + "$", value, opts);
		},
		removeMeta(key, opts = {}) {
			return this.removeItem(key + "$", opts);
		},
		async getKeys(base, opts = {}) {
			base = normalizeBaseKey(base);
			const mounts = getMounts(base, true);
			let maskedMounts = [];
			const allKeys = [];
			let allMountsSupportMaxDepth = true;
			for (const mount of mounts) {
				if (!mount.driver.flags?.maxDepth) allMountsSupportMaxDepth = false;
				const rawKeys = await asyncCall(mount.driver.getKeys, mount.relativeBase, opts);
				for (const key of rawKeys) {
					const fullKey = mount.mountpoint + normalizeKey(key);
					if (!maskedMounts.some((p) => fullKey.startsWith(p))) allKeys.push(fullKey);
				}
				maskedMounts = [mount.mountpoint, ...maskedMounts.filter((p) => !p.startsWith(mount.mountpoint))];
			}
			const shouldFilterByDepth = opts.maxDepth !== void 0 && !allMountsSupportMaxDepth;
			return allKeys.filter((key) => (!shouldFilterByDepth || filterKeyByDepth(key, opts.maxDepth)) && filterKeyByBase(key, base));
		},
		async clear(base, opts = {}) {
			base = normalizeBaseKey(base);
			await Promise.all(getMounts(base, false).map(async (m) => {
				if (m.driver.clear) return asyncCall(m.driver.clear, m.relativeBase, opts);
				if (m.driver.removeItem) {
					const keys = await m.driver.getKeys(m.relativeBase || "", opts);
					return Promise.all(keys.map((key) => m.driver.removeItem(key, opts)));
				}
			}));
		},
		async dispose() {
			await Promise.all(Object.values(context.mounts).map((driver) => dispose(driver)));
		},
		async watch(callback) {
			await startWatch();
			context.watchListeners.push(callback);
			return async () => {
				context.watchListeners = context.watchListeners.filter((listener) => listener !== callback);
				if (context.watchListeners.length === 0) await stopWatch();
			};
		},
		async unwatch() {
			context.watchListeners = [];
			await stopWatch();
		},
		mount(base, driver) {
			base = normalizeBaseKey(base);
			if (base && context.mounts[base]) throw new Error(`already mounted at ${base}`);
			if (base) {
				context.mountpoints.push(base);
				context.mountpoints.sort((a, b) => b.length - a.length);
			}
			context.mounts[base] = driver;
			if (context.watching) Promise.resolve(watch(driver, onChange, base)).then((unwatcher) => {
				context.unwatch[base] = unwatcher;
			}).catch(console.error);
			return storage;
		},
		async unmount(base, _dispose = true) {
			base = normalizeBaseKey(base);
			if (!base || !context.mounts[base]) return;
			if (context.watching && base in context.unwatch) {
				context.unwatch[base]?.();
				delete context.unwatch[base];
			}
			if (_dispose) await dispose(context.mounts[base]);
			context.mountpoints = context.mountpoints.filter((key) => key !== base);
			delete context.mounts[base];
		},
		getMount(key = "") {
			key = normalizeKey(key) + ":";
			const m = getMount(key);
			return {
				driver: m.driver,
				base: m.base
			};
		},
		getMounts(base = "", opts = {}) {
			base = normalizeKey(base);
			const mounts = getMounts(base, opts.parents);
			return mounts.map((m) => ({
				driver: m.driver,
				base: m.mountpoint
			}));
		},
		keys: (base, opts = {}) => storage.getKeys(base, opts),
		get: (key, opts = {}) => storage.getItem(key, opts),
		set: (key, value, opts = {}) => storage.setItem(key, value, opts),
		has: (key, opts = {}) => storage.hasItem(key, opts),
		del: (key, opts = {}) => storage.removeItem(key, opts),
		remove: (key, opts = {}) => storage.removeItem(key, opts)
	};
	return storage;
}
function watch(driver, onChange, base) {
	return driver.watch ? driver.watch((event, key) => onChange(event, base + key)) : () => {};
}
async function dispose(driver) {
	if (typeof driver.dispose === "function") await asyncCall(driver.dispose);
}

//#endregion
//#region node_modules/unstorage/drivers/utils/index.mjs
function defineDriver(factory) {
	return factory;
}
function createError(driver, message, opts) {
	const err = new Error(`[unstorage] [${driver}] ${message}`, opts);
	if (Error.captureStackTrace) Error.captureStackTrace(err, createError);
	return err;
}
function createRequiredError(driver, name) {
	if (Array.isArray(name)) return createError(driver, `Missing some of the required options ${name.map((n) => "`" + n + "`").join(", ")}`);
	return createError(driver, `Missing required option \`${name}\`.`);
}

//#endregion
//#region node_modules/unstorage/drivers/utils/node-fs.mjs
function ignoreNotfound(err) {
	return err.code === "ENOENT" || err.code === "EISDIR" ? null : err;
}
function ignoreExists(err) {
	return err.code === "EEXIST" ? null : err;
}
async function writeFile(path$1, data, encoding) {
	await ensuredir(dirname(path$1));
	return promises.writeFile(path$1, data, encoding);
}
function readFile(path$1, encoding) {
	return promises.readFile(path$1, encoding).catch(ignoreNotfound);
}
function unlink(path$1) {
	return promises.unlink(path$1).catch(ignoreNotfound);
}
function readdir(dir) {
	return promises.readdir(dir, { withFileTypes: true }).catch(ignoreNotfound).then((r) => r || []);
}
async function ensuredir(dir) {
	if (existsSync(dir)) return;
	await ensuredir(dirname(dir)).catch(ignoreExists);
	await promises.mkdir(dir).catch(ignoreExists);
}
async function readdirRecursive(dir, ignore, maxDepth) {
	if (ignore && ignore(dir)) return [];
	const entries = await readdir(dir);
	const files = [];
	await Promise.all(entries.map(async (entry) => {
		const entryPath = resolve(dir, entry.name);
		if (entry.isDirectory()) {
			if (maxDepth === void 0 || maxDepth > 0) {
				const dirFiles = await readdirRecursive(entryPath, ignore, maxDepth === void 0 ? void 0 : maxDepth - 1);
				files.push(...dirFiles.map((f) => entry.name + "/" + f));
			}
		} else if (!(ignore && ignore(entry.name))) files.push(entry.name);
	}));
	return files;
}
async function rmRecursive(dir) {
	const entries = await readdir(dir);
	await Promise.all(entries.map((entry) => {
		const entryPath = resolve(dir, entry.name);
		if (entry.isDirectory()) return rmRecursive(entryPath).then(() => promises.rmdir(entryPath));
		else return promises.unlink(entryPath);
	}));
}

//#endregion
//#region node_modules/unstorage/drivers/fs-lite.mjs
const PATH_TRAVERSE_RE = /\.\.:|\.\.$/;
const DRIVER_NAME = "fs-lite";
var fs_lite_default = defineDriver((opts = {}) => {
	if (!opts.base) throw createRequiredError(DRIVER_NAME, "base");
	opts.base = resolve(opts.base);
	const r = (key) => {
		if (PATH_TRAVERSE_RE.test(key)) throw createError(DRIVER_NAME, `Invalid key: ${JSON.stringify(key)}. It should not contain .. segments`);
		const resolved = join(opts.base, key.replace(/:/g, "/"));
		return resolved;
	};
	return {
		name: DRIVER_NAME,
		options: opts,
		flags: { maxDepth: true },
		hasItem(key) {
			return existsSync(r(key));
		},
		getItem(key) {
			return readFile(r(key), "utf8");
		},
		getItemRaw(key) {
			return readFile(r(key));
		},
		async getMeta(key) {
			const { atime, mtime, size, birthtime, ctime } = await promises.stat(r(key)).catch(() => ({}));
			return {
				atime,
				mtime,
				size,
				birthtime,
				ctime
			};
		},
		setItem(key, value) {
			if (opts.readOnly) return;
			return writeFile(r(key), value, "utf8");
		},
		setItemRaw(key, value) {
			if (opts.readOnly) return;
			return writeFile(r(key), value);
		},
		removeItem(key) {
			if (opts.readOnly) return;
			return unlink(r(key));
		},
		getKeys(_base, topts) {
			return readdirRecursive(r("."), opts.ignore, topts?.maxDepth);
		},
		async clear() {
			if (opts.readOnly || opts.noClear) return;
			await rmRecursive(r("."));
		}
	};
});

//#endregion
//#region node_modules/nanoid/non-secure/index.js
let customAlphabet = (alphabet, defaultSize = 21) => {
	return (size = defaultSize) => {
		let id = "";
		let i = size | 0;
		while (i--) id += alphabet[Math.random() * alphabet.length | 0];
		return id;
	};
};

//#endregion
//#region node_modules/secure-json-parse/index.js
var require_secure_json_parse = __commonJS({ "node_modules/secure-json-parse/index.js"(exports, module) {
	const hasBuffer = typeof Buffer !== "undefined";
	const suspectProtoRx = /"(?:_|\\u005[Ff])(?:_|\\u005[Ff])(?:p|\\u0070)(?:r|\\u0072)(?:o|\\u006[Ff])(?:t|\\u0074)(?:o|\\u006[Ff])(?:_|\\u005[Ff])(?:_|\\u005[Ff])"\s*:/;
	const suspectConstructorRx = /"(?:c|\\u0063)(?:o|\\u006[Ff])(?:n|\\u006[Ee])(?:s|\\u0073)(?:t|\\u0074)(?:r|\\u0072)(?:u|\\u0075)(?:c|\\u0063)(?:t|\\u0074)(?:o|\\u006[Ff])(?:r|\\u0072)"\s*:/;
	function _parse(text, reviver, options) {
		if (options == null) {
			if (reviver !== null && typeof reviver === "object") {
				options = reviver;
				reviver = void 0;
			}
		}
		if (hasBuffer && Buffer.isBuffer(text)) text = text.toString();
		if (text && text.charCodeAt(0) === 65279) text = text.slice(1);
		const obj = JSON.parse(text, reviver);
		if (obj === null || typeof obj !== "object") return obj;
		const protoAction = options && options.protoAction || "error";
		const constructorAction = options && options.constructorAction || "error";
		if (protoAction === "ignore" && constructorAction === "ignore") return obj;
		if (protoAction !== "ignore" && constructorAction !== "ignore") {
			if (suspectProtoRx.test(text) === false && suspectConstructorRx.test(text) === false) return obj;
		} else if (protoAction !== "ignore" && constructorAction === "ignore") {
			if (suspectProtoRx.test(text) === false) return obj;
		} else if (suspectConstructorRx.test(text) === false) return obj;
		return filter(obj, {
			protoAction,
			constructorAction,
			safe: options && options.safe
		});
	}
	function filter(obj, { protoAction = "error", constructorAction = "error", safe } = {}) {
		let next = [obj];
		while (next.length) {
			const nodes = next;
			next = [];
			for (const node of nodes) {
				if (protoAction !== "ignore" && Object.prototype.hasOwnProperty.call(node, "__proto__")) {
					if (safe === true) return null;
					else if (protoAction === "error") throw new SyntaxError("Object contains forbidden prototype property");
					delete node.__proto__;
				}
				if (constructorAction !== "ignore" && Object.prototype.hasOwnProperty.call(node, "constructor") && Object.prototype.hasOwnProperty.call(node.constructor, "prototype")) {
					if (safe === true) return null;
					else if (constructorAction === "error") throw new SyntaxError("Object contains forbidden prototype property");
					delete node.constructor;
				}
				for (const key in node) {
					const value = node[key];
					if (value && typeof value === "object") next.push(value);
				}
			}
		}
		return obj;
	}
	function parse(text, reviver, options) {
		const stackTraceLimit = Error.stackTraceLimit;
		Error.stackTraceLimit = 0;
		try {
			return _parse(text, reviver, options);
		} finally {
			Error.stackTraceLimit = stackTraceLimit;
		}
	}
	function safeParse(text, reviver) {
		const stackTraceLimit = Error.stackTraceLimit;
		Error.stackTraceLimit = 0;
		try {
			return _parse(text, reviver, { safe: true });
		} catch (_e) {
			return null;
		} finally {
			Error.stackTraceLimit = stackTraceLimit;
		}
	}
	module.exports = parse;
	module.exports.default = parse;
	module.exports.parse = parse;
	module.exports.safeParse = safeParse;
	module.exports.scan = filter;
} });
var import_secure_json_parse = __toESM(require_secure_json_parse(), 1);

//#endregion
//#region node_modules/@ai-sdk/provider-utils/dist/index.mjs
function combineHeaders(...headers) {
	return headers.reduce((combinedHeaders, currentHeaders) => ({
		...combinedHeaders,
		...currentHeaders != null ? currentHeaders : {}
	}), {});
}
function createEventSourceParserStream() {
	let buffer = "";
	let event = void 0;
	let data = [];
	let lastEventId = void 0;
	let retry = void 0;
	function parseLine(line, controller) {
		if (line === "") {
			dispatchEvent(controller);
			return;
		}
		if (line.startsWith(":")) return;
		const colonIndex = line.indexOf(":");
		if (colonIndex === -1) {
			handleField(line, "");
			return;
		}
		const field = line.slice(0, colonIndex);
		const valueStart = colonIndex + 1;
		const value = valueStart < line.length && line[valueStart] === " " ? line.slice(valueStart + 1) : line.slice(valueStart);
		handleField(field, value);
	}
	function dispatchEvent(controller) {
		if (data.length > 0) {
			controller.enqueue({
				event,
				data: data.join("\n"),
				id: lastEventId,
				retry
			});
			data = [];
			event = void 0;
			retry = void 0;
		}
	}
	function handleField(field, value) {
		switch (field) {
			case "event":
				event = value;
				break;
			case "data":
				data.push(value);
				break;
			case "id":
				lastEventId = value;
				break;
			case "retry":
				const parsedRetry = parseInt(value, 10);
				if (!isNaN(parsedRetry)) retry = parsedRetry;
				break;
		}
	}
	return new TransformStream({
		transform(chunk, controller) {
			const { lines, incompleteLine } = splitLines(buffer, chunk);
			buffer = incompleteLine;
			for (let i = 0; i < lines.length; i++) parseLine(lines[i], controller);
		},
		flush(controller) {
			parseLine(buffer, controller);
			dispatchEvent(controller);
		}
	});
}
function splitLines(buffer, chunk) {
	const lines = [];
	let currentLine = buffer;
	for (let i = 0; i < chunk.length;) {
		const char = chunk[i++];
		if (char === "\n") {
			lines.push(currentLine);
			currentLine = "";
		} else if (char === "\r") {
			lines.push(currentLine);
			currentLine = "";
			if (chunk[i + 1] === "\n") i++;
		} else currentLine += char;
	}
	return {
		lines,
		incompleteLine: currentLine
	};
}
function extractResponseHeaders(response) {
	const headers = {};
	response.headers.forEach((value, key) => {
		headers[key] = value;
	});
	return headers;
}
var createIdGenerator = ({ prefix, size: defaultSize = 16, alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", separator = "-" } = {}) => {
	const generator = customAlphabet(alphabet, defaultSize);
	if (prefix == null) return generator;
	if (alphabet.includes(separator)) throw new InvalidArgumentError({
		argument: "separator",
		message: `The separator "${separator}" must not be part of the alphabet "${alphabet}".`
	});
	return (size) => `${prefix}${separator}${generator(size)}`;
};
var generateId = createIdGenerator();
function removeUndefinedEntries(record) {
	return Object.fromEntries(Object.entries(record).filter(([_key, value]) => value != null));
}
function isAbortError(error) {
	return error instanceof Error && (error.name === "AbortError" || error.name === "TimeoutError");
}
var validatorSymbol = Symbol.for("vercel.ai.validator");
function validator(validate) {
	return {
		[validatorSymbol]: true,
		validate
	};
}
function isValidator(value) {
	return typeof value === "object" && value !== null && validatorSymbol in value && value[validatorSymbol] === true && "validate" in value;
}
function asValidator(value) {
	return isValidator(value) ? value : zodValidator(value);
}
function zodValidator(zodSchema) {
	return validator((value) => {
		const result = zodSchema.safeParse(value);
		return result.success ? {
			success: true,
			value: result.data
		} : {
			success: false,
			error: result.error
		};
	});
}
function validateTypes({ value, schema: inputSchema }) {
	const result = safeValidateTypes({
		value,
		schema: inputSchema
	});
	if (!result.success) throw TypeValidationError.wrap({
		value,
		cause: result.error
	});
	return result.value;
}
function safeValidateTypes({ value, schema }) {
	const validator2 = asValidator(schema);
	try {
		if (validator2.validate == null) return {
			success: true,
			value
		};
		const result = validator2.validate(value);
		if (result.success) return result;
		return {
			success: false,
			error: TypeValidationError.wrap({
				value,
				cause: result.error
			})
		};
	} catch (error) {
		return {
			success: false,
			error: TypeValidationError.wrap({
				value,
				cause: error
			})
		};
	}
}
function parseJSON({ text, schema }) {
	try {
		const value = import_secure_json_parse.default.parse(text);
		if (schema == null) return value;
		return validateTypes({
			value,
			schema
		});
	} catch (error) {
		if (JSONParseError.isInstance(error) || TypeValidationError.isInstance(error)) throw error;
		throw new JSONParseError({
			text,
			cause: error
		});
	}
}
function safeParseJSON({ text, schema }) {
	try {
		const value = import_secure_json_parse.default.parse(text);
		if (schema == null) return {
			success: true,
			value,
			rawValue: value
		};
		const validationResult = safeValidateTypes({
			value,
			schema
		});
		return validationResult.success ? {
			...validationResult,
			rawValue: value
		} : validationResult;
	} catch (error) {
		return {
			success: false,
			error: JSONParseError.isInstance(error) ? error : new JSONParseError({
				text,
				cause: error
			})
		};
	}
}
function isParsableJson(input) {
	try {
		import_secure_json_parse.default.parse(input);
		return true;
	} catch (e) {
		return false;
	}
}
var getOriginalFetch2 = () => globalThis.fetch;
var postJsonToApi = async ({ url, headers, body, failedResponseHandler, successfulResponseHandler, abortSignal, fetch }) => postToApi({
	url,
	headers: {
		"Content-Type": "application/json",
		...headers
	},
	body: {
		content: JSON.stringify(body),
		values: body
	},
	failedResponseHandler,
	successfulResponseHandler,
	abortSignal,
	fetch
});
var postToApi = async ({ url, headers = {}, body, successfulResponseHandler, failedResponseHandler, abortSignal, fetch = getOriginalFetch2() }) => {
	try {
		const response = await fetch(url, {
			method: "POST",
			headers: removeUndefinedEntries(headers),
			body: body.content,
			signal: abortSignal
		});
		const responseHeaders = extractResponseHeaders(response);
		if (!response.ok) {
			let errorInformation;
			try {
				errorInformation = await failedResponseHandler({
					response,
					url,
					requestBodyValues: body.values
				});
			} catch (error) {
				if (isAbortError(error) || APICallError.isInstance(error)) throw error;
				throw new APICallError({
					message: "Failed to process error response",
					cause: error,
					statusCode: response.status,
					url,
					responseHeaders,
					requestBodyValues: body.values
				});
			}
			throw errorInformation.value;
		}
		try {
			return await successfulResponseHandler({
				response,
				url,
				requestBodyValues: body.values
			});
		} catch (error) {
			if (error instanceof Error) {
				if (isAbortError(error) || APICallError.isInstance(error)) throw error;
			}
			throw new APICallError({
				message: "Failed to process successful response",
				cause: error,
				statusCode: response.status,
				url,
				responseHeaders,
				requestBodyValues: body.values
			});
		}
	} catch (error) {
		if (isAbortError(error)) throw error;
		if (error instanceof TypeError && error.message === "fetch failed") {
			const cause = error.cause;
			if (cause != null) throw new APICallError({
				message: `Cannot connect to API: ${cause.message}`,
				cause,
				url,
				requestBodyValues: body.values,
				isRetryable: true
			});
		}
		throw error;
	}
};
var createJsonErrorResponseHandler = ({ errorSchema, errorToMessage, isRetryable }) => async ({ response, url, requestBodyValues }) => {
	const responseBody = await response.text();
	const responseHeaders = extractResponseHeaders(response);
	if (responseBody.trim() === "") return {
		responseHeaders,
		value: new APICallError({
			message: response.statusText,
			url,
			requestBodyValues,
			statusCode: response.status,
			responseHeaders,
			responseBody,
			isRetryable: isRetryable == null ? void 0 : isRetryable(response)
		})
	};
	try {
		const parsedError = parseJSON({
			text: responseBody,
			schema: errorSchema
		});
		return {
			responseHeaders,
			value: new APICallError({
				message: errorToMessage(parsedError),
				url,
				requestBodyValues,
				statusCode: response.status,
				responseHeaders,
				responseBody,
				data: parsedError,
				isRetryable: isRetryable == null ? void 0 : isRetryable(response, parsedError)
			})
		};
	} catch (parseError) {
		return {
			responseHeaders,
			value: new APICallError({
				message: response.statusText,
				url,
				requestBodyValues,
				statusCode: response.status,
				responseHeaders,
				responseBody,
				isRetryable: isRetryable == null ? void 0 : isRetryable(response)
			})
		};
	}
};
var createEventSourceResponseHandler = (chunkSchema) => async ({ response }) => {
	const responseHeaders = extractResponseHeaders(response);
	if (response.body == null) throw new EmptyResponseBodyError({});
	return {
		responseHeaders,
		value: response.body.pipeThrough(new TextDecoderStream()).pipeThrough(createEventSourceParserStream()).pipeThrough(new TransformStream({ transform({ data }, controller) {
			if (data === "[DONE]") return;
			controller.enqueue(safeParseJSON({
				text: data,
				schema: chunkSchema
			}));
		} }))
	};
};
var createJsonResponseHandler = (responseSchema) => async ({ response, url, requestBodyValues }) => {
	const responseBody = await response.text();
	const parsedResult = safeParseJSON({
		text: responseBody,
		schema: responseSchema
	});
	const responseHeaders = extractResponseHeaders(response);
	if (!parsedResult.success) throw new APICallError({
		message: "Invalid JSON response",
		cause: parsedResult.error,
		statusCode: response.status,
		responseHeaders,
		responseBody,
		url,
		requestBodyValues
	});
	return {
		responseHeaders,
		value: parsedResult.value,
		rawValue: parsedResult.rawValue
	};
};
var { btoa, atob } = globalThis;
function convertUint8ArrayToBase64(array) {
	let latin1string = "";
	for (let i = 0; i < array.length; i++) latin1string += String.fromCodePoint(array[i]);
	return btoa(latin1string);
}
function withoutTrailingSlash(url) {
	return url == null ? void 0 : url.replace(/\/$/, "");
}

//#endregion
//#region node_modules/@ai-sdk/openai-compatible/dist/index.mjs
function getOpenAIMetadata(message) {
	var _a, _b;
	return (_b = (_a = message == null ? void 0 : message.providerMetadata) == null ? void 0 : _a.openaiCompatible) != null ? _b : {};
}
function convertToOpenAICompatibleChatMessages(prompt) {
	const messages = [];
	for (const { role, content,...message } of prompt) {
		const metadata = getOpenAIMetadata({ ...message });
		switch (role) {
			case "system": {
				messages.push({
					role: "system",
					content,
					...metadata
				});
				break;
			}
			case "user": {
				if (content.length === 1 && content[0].type === "text") {
					messages.push({
						role: "user",
						content: content[0].text,
						...getOpenAIMetadata(content[0])
					});
					break;
				}
				messages.push({
					role: "user",
					content: content.map((part) => {
						var _a;
						const partMetadata = getOpenAIMetadata(part);
						switch (part.type) {
							case "text": return {
								type: "text",
								text: part.text,
								...partMetadata
							};
							case "image": return {
								type: "image_url",
								image_url: { url: part.image instanceof URL ? part.image.toString() : `data:${(_a = part.mimeType) != null ? _a : "image/jpeg"};base64,${convertUint8ArrayToBase64(part.image)}` },
								...partMetadata
							};
							case "file": throw new UnsupportedFunctionalityError({ functionality: "File content parts in user messages" });
						}
					}),
					...metadata
				});
				break;
			}
			case "assistant": {
				let text = "";
				const toolCalls = [];
				for (const part of content) {
					const partMetadata = getOpenAIMetadata(part);
					switch (part.type) {
						case "text": {
							text += part.text;
							break;
						}
						case "tool-call": {
							toolCalls.push({
								id: part.toolCallId,
								type: "function",
								function: {
									name: part.toolName,
									arguments: JSON.stringify(part.args)
								},
								...partMetadata
							});
							break;
						}
					}
				}
				messages.push({
					role: "assistant",
					content: text,
					tool_calls: toolCalls.length > 0 ? toolCalls : void 0,
					...metadata
				});
				break;
			}
			case "tool": {
				for (const toolResponse of content) {
					const toolResponseMetadata = getOpenAIMetadata(toolResponse);
					messages.push({
						role: "tool",
						tool_call_id: toolResponse.toolCallId,
						content: JSON.stringify(toolResponse.result),
						...toolResponseMetadata
					});
				}
				break;
			}
			default: {
				const _exhaustiveCheck = role;
				throw new Error(`Unsupported role: ${_exhaustiveCheck}`);
			}
		}
	}
	return messages;
}
function getResponseMetadata({ id, model, created }) {
	return {
		id: id != null ? id : void 0,
		modelId: model != null ? model : void 0,
		timestamp: created != null ? new Date(created * 1e3) : void 0
	};
}
function mapOpenAICompatibleFinishReason(finishReason) {
	switch (finishReason) {
		case "stop": return "stop";
		case "length": return "length";
		case "content_filter": return "content-filter";
		case "function_call":
		case "tool_calls": return "tool-calls";
		default: return "unknown";
	}
}
var openaiCompatibleErrorDataSchema = z.object({ error: z.object({
	message: z.string(),
	type: z.string().nullish(),
	param: z.any().nullish(),
	code: z.union([z.string(), z.number()]).nullish()
}) });
var defaultOpenAICompatibleErrorStructure = {
	errorSchema: openaiCompatibleErrorDataSchema,
	errorToMessage: (data) => data.error.message
};
function prepareTools({ mode, structuredOutputs }) {
	var _a;
	const tools = ((_a = mode.tools) == null ? void 0 : _a.length) ? mode.tools : void 0;
	const toolWarnings = [];
	if (tools == null) return {
		tools: void 0,
		tool_choice: void 0,
		toolWarnings
	};
	const toolChoice = mode.toolChoice;
	const openaiCompatTools = [];
	for (const tool of tools) if (tool.type === "provider-defined") toolWarnings.push({
		type: "unsupported-tool",
		tool
	});
	else openaiCompatTools.push({
		type: "function",
		function: {
			name: tool.name,
			description: tool.description,
			parameters: tool.parameters
		}
	});
	if (toolChoice == null) return {
		tools: openaiCompatTools,
		tool_choice: void 0,
		toolWarnings
	};
	const type = toolChoice.type;
	switch (type) {
		case "auto":
		case "none":
		case "required": return {
			tools: openaiCompatTools,
			tool_choice: type,
			toolWarnings
		};
		case "tool": return {
			tools: openaiCompatTools,
			tool_choice: {
				type: "function",
				function: { name: toolChoice.toolName }
			},
			toolWarnings
		};
		default: {
			const _exhaustiveCheck = type;
			throw new UnsupportedFunctionalityError({ functionality: `Unsupported tool choice type: ${_exhaustiveCheck}` });
		}
	}
}
var OpenAICompatibleChatLanguageModel = class {
	constructor(modelId, settings, config) {
		this.specificationVersion = "v1";
		var _a, _b;
		this.modelId = modelId;
		this.settings = settings;
		this.config = config;
		const errorStructure = (_a = config.errorStructure) != null ? _a : defaultOpenAICompatibleErrorStructure;
		this.chunkSchema = createOpenAICompatibleChatChunkSchema(errorStructure.errorSchema);
		this.failedResponseHandler = createJsonErrorResponseHandler(errorStructure);
		this.supportsStructuredOutputs = (_b = config.supportsStructuredOutputs) != null ? _b : false;
	}
	get defaultObjectGenerationMode() {
		return this.config.defaultObjectGenerationMode;
	}
	get provider() {
		return this.config.provider;
	}
	get providerOptionsName() {
		return this.config.provider.split(".")[0].trim();
	}
	getArgs({ mode, prompt, maxTokens, temperature, topP, topK, frequencyPenalty, presencePenalty, providerMetadata, stopSequences, responseFormat, seed }) {
		var _a, _b;
		const type = mode.type;
		const warnings = [];
		if (topK != null) warnings.push({
			type: "unsupported-setting",
			setting: "topK"
		});
		if ((responseFormat == null ? void 0 : responseFormat.type) === "json" && responseFormat.schema != null && !this.supportsStructuredOutputs) warnings.push({
			type: "unsupported-setting",
			setting: "responseFormat",
			details: "JSON response format schema is only supported with structuredOutputs"
		});
		const baseArgs = {
			model: this.modelId,
			user: this.settings.user,
			max_tokens: maxTokens,
			temperature,
			top_p: topP,
			frequency_penalty: frequencyPenalty,
			presence_penalty: presencePenalty,
			response_format: (responseFormat == null ? void 0 : responseFormat.type) === "json" ? this.supportsStructuredOutputs === true && responseFormat.schema != null ? {
				type: "json_schema",
				json_schema: {
					schema: responseFormat.schema,
					name: (_a = responseFormat.name) != null ? _a : "response",
					description: responseFormat.description
				}
			} : { type: "json_object" } : void 0,
			stop: stopSequences,
			seed,
			...providerMetadata == null ? void 0 : providerMetadata[this.providerOptionsName],
			messages: convertToOpenAICompatibleChatMessages(prompt)
		};
		switch (type) {
			case "regular": {
				const { tools, tool_choice, toolWarnings } = prepareTools({
					mode,
					structuredOutputs: this.supportsStructuredOutputs
				});
				return {
					args: {
						...baseArgs,
						tools,
						tool_choice
					},
					warnings: [...warnings, ...toolWarnings]
				};
			}
			case "object-json": return {
				args: {
					...baseArgs,
					response_format: this.supportsStructuredOutputs === true && mode.schema != null ? {
						type: "json_schema",
						json_schema: {
							schema: mode.schema,
							name: (_b = mode.name) != null ? _b : "response",
							description: mode.description
						}
					} : { type: "json_object" }
				},
				warnings
			};
			case "object-tool": return {
				args: {
					...baseArgs,
					tool_choice: {
						type: "function",
						function: { name: mode.tool.name }
					},
					tools: [{
						type: "function",
						function: {
							name: mode.tool.name,
							description: mode.tool.description,
							parameters: mode.tool.parameters
						}
					}]
				},
				warnings
			};
			default: {
				const _exhaustiveCheck = type;
				throw new Error(`Unsupported type: ${_exhaustiveCheck}`);
			}
		}
	}
	async doGenerate(options) {
		var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k;
		const { args, warnings } = this.getArgs({ ...options });
		const body = JSON.stringify(args);
		const { responseHeaders, value: responseBody, rawValue: rawResponse } = await postJsonToApi({
			url: this.config.url({
				path: "/chat/completions",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), options.headers),
			body: args,
			failedResponseHandler: this.failedResponseHandler,
			successfulResponseHandler: createJsonResponseHandler(OpenAICompatibleChatResponseSchema),
			abortSignal: options.abortSignal,
			fetch: this.config.fetch
		});
		const { messages: rawPrompt,...rawSettings } = args;
		const choice = responseBody.choices[0];
		const providerMetadata = {
			[this.providerOptionsName]: {},
			...(_b = (_a = this.config.metadataExtractor) == null ? void 0 : _a.extractMetadata) == null ? void 0 : _b.call(_a, { parsedBody: rawResponse })
		};
		const completionTokenDetails = (_c = responseBody.usage) == null ? void 0 : _c.completion_tokens_details;
		const promptTokenDetails = (_d = responseBody.usage) == null ? void 0 : _d.prompt_tokens_details;
		if ((completionTokenDetails == null ? void 0 : completionTokenDetails.reasoning_tokens) != null) providerMetadata[this.providerOptionsName].reasoningTokens = completionTokenDetails == null ? void 0 : completionTokenDetails.reasoning_tokens;
		if ((completionTokenDetails == null ? void 0 : completionTokenDetails.accepted_prediction_tokens) != null) providerMetadata[this.providerOptionsName].acceptedPredictionTokens = completionTokenDetails == null ? void 0 : completionTokenDetails.accepted_prediction_tokens;
		if ((completionTokenDetails == null ? void 0 : completionTokenDetails.rejected_prediction_tokens) != null) providerMetadata[this.providerOptionsName].rejectedPredictionTokens = completionTokenDetails == null ? void 0 : completionTokenDetails.rejected_prediction_tokens;
		if ((promptTokenDetails == null ? void 0 : promptTokenDetails.cached_tokens) != null) providerMetadata[this.providerOptionsName].cachedPromptTokens = promptTokenDetails == null ? void 0 : promptTokenDetails.cached_tokens;
		return {
			text: (_e = choice.message.content) != null ? _e : void 0,
			reasoning: (_f = choice.message.reasoning_content) != null ? _f : void 0,
			toolCalls: (_g = choice.message.tool_calls) == null ? void 0 : _g.map((toolCall) => {
				var _a2;
				return {
					toolCallType: "function",
					toolCallId: (_a2 = toolCall.id) != null ? _a2 : generateId(),
					toolName: toolCall.function.name,
					args: toolCall.function.arguments
				};
			}),
			finishReason: mapOpenAICompatibleFinishReason(choice.finish_reason),
			usage: {
				promptTokens: (_i = (_h = responseBody.usage) == null ? void 0 : _h.prompt_tokens) != null ? _i : NaN,
				completionTokens: (_k = (_j = responseBody.usage) == null ? void 0 : _j.completion_tokens) != null ? _k : NaN
			},
			providerMetadata,
			rawCall: {
				rawPrompt,
				rawSettings
			},
			rawResponse: {
				headers: responseHeaders,
				body: rawResponse
			},
			response: getResponseMetadata(responseBody),
			warnings,
			request: { body }
		};
	}
	async doStream(options) {
		var _a;
		if (this.settings.simulateStreaming) {
			const result = await this.doGenerate(options);
			const simulatedStream = new ReadableStream({ start(controller) {
				controller.enqueue({
					type: "response-metadata",
					...result.response
				});
				if (result.reasoning) if (Array.isArray(result.reasoning)) {
					for (const part of result.reasoning) if (part.type === "text") controller.enqueue({
						type: "reasoning",
						textDelta: part.text
					});
				} else controller.enqueue({
					type: "reasoning",
					textDelta: result.reasoning
				});
				if (result.text) controller.enqueue({
					type: "text-delta",
					textDelta: result.text
				});
				if (result.toolCalls) for (const toolCall of result.toolCalls) controller.enqueue({
					type: "tool-call",
					...toolCall
				});
				controller.enqueue({
					type: "finish",
					finishReason: result.finishReason,
					usage: result.usage,
					logprobs: result.logprobs,
					providerMetadata: result.providerMetadata
				});
				controller.close();
			} });
			return {
				stream: simulatedStream,
				rawCall: result.rawCall,
				rawResponse: result.rawResponse,
				warnings: result.warnings
			};
		}
		const { args, warnings } = this.getArgs({ ...options });
		const body = {
			...args,
			stream: true,
			stream_options: this.config.includeUsage ? { include_usage: true } : void 0
		};
		const metadataExtractor = (_a = this.config.metadataExtractor) == null ? void 0 : _a.createStreamExtractor();
		const { responseHeaders, value: response } = await postJsonToApi({
			url: this.config.url({
				path: "/chat/completions",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), options.headers),
			body,
			failedResponseHandler: this.failedResponseHandler,
			successfulResponseHandler: createEventSourceResponseHandler(this.chunkSchema),
			abortSignal: options.abortSignal,
			fetch: this.config.fetch
		});
		const { messages: rawPrompt,...rawSettings } = args;
		const toolCalls = [];
		let finishReason = "unknown";
		let usage = {
			completionTokens: void 0,
			completionTokensDetails: {
				reasoningTokens: void 0,
				acceptedPredictionTokens: void 0,
				rejectedPredictionTokens: void 0
			},
			promptTokens: void 0,
			promptTokensDetails: { cachedTokens: void 0 }
		};
		let isFirstChunk = true;
		let providerOptionsName = this.providerOptionsName;
		return {
			stream: response.pipeThrough(new TransformStream({
				transform(chunk, controller) {
					var _a2, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l;
					if (!chunk.success) {
						finishReason = "error";
						controller.enqueue({
							type: "error",
							error: chunk.error
						});
						return;
					}
					const value = chunk.value;
					metadataExtractor == null ? void 0 : metadataExtractor.processChunk(chunk.rawValue);
					if ("error" in value) {
						finishReason = "error";
						controller.enqueue({
							type: "error",
							error: value.error.message
						});
						return;
					}
					if (isFirstChunk) {
						isFirstChunk = false;
						controller.enqueue({
							type: "response-metadata",
							...getResponseMetadata(value)
						});
					}
					if (value.usage != null) {
						const { prompt_tokens, completion_tokens, prompt_tokens_details, completion_tokens_details } = value.usage;
						usage.promptTokens = prompt_tokens != null ? prompt_tokens : void 0;
						usage.completionTokens = completion_tokens != null ? completion_tokens : void 0;
						if ((completion_tokens_details == null ? void 0 : completion_tokens_details.reasoning_tokens) != null) usage.completionTokensDetails.reasoningTokens = completion_tokens_details == null ? void 0 : completion_tokens_details.reasoning_tokens;
						if ((completion_tokens_details == null ? void 0 : completion_tokens_details.accepted_prediction_tokens) != null) usage.completionTokensDetails.acceptedPredictionTokens = completion_tokens_details == null ? void 0 : completion_tokens_details.accepted_prediction_tokens;
						if ((completion_tokens_details == null ? void 0 : completion_tokens_details.rejected_prediction_tokens) != null) usage.completionTokensDetails.rejectedPredictionTokens = completion_tokens_details == null ? void 0 : completion_tokens_details.rejected_prediction_tokens;
						if ((prompt_tokens_details == null ? void 0 : prompt_tokens_details.cached_tokens) != null) usage.promptTokensDetails.cachedTokens = prompt_tokens_details == null ? void 0 : prompt_tokens_details.cached_tokens;
					}
					const choice = value.choices[0];
					if ((choice == null ? void 0 : choice.finish_reason) != null) finishReason = mapOpenAICompatibleFinishReason(choice.finish_reason);
					if ((choice == null ? void 0 : choice.delta) == null) return;
					const delta = choice.delta;
					if (delta.reasoning_content != null) controller.enqueue({
						type: "reasoning",
						textDelta: delta.reasoning_content
					});
					if (delta.content != null) controller.enqueue({
						type: "text-delta",
						textDelta: delta.content
					});
					if (delta.tool_calls != null) for (const toolCallDelta of delta.tool_calls) {
						const index = toolCallDelta.index;
						if (toolCalls[index] == null) {
							if (toolCallDelta.type !== "function") throw new InvalidResponseDataError({
								data: toolCallDelta,
								message: `Expected 'function' type.`
							});
							if (toolCallDelta.id == null) throw new InvalidResponseDataError({
								data: toolCallDelta,
								message: `Expected 'id' to be a string.`
							});
							if (((_a2 = toolCallDelta.function) == null ? void 0 : _a2.name) == null) throw new InvalidResponseDataError({
								data: toolCallDelta,
								message: `Expected 'function.name' to be a string.`
							});
							toolCalls[index] = {
								id: toolCallDelta.id,
								type: "function",
								function: {
									name: toolCallDelta.function.name,
									arguments: (_b = toolCallDelta.function.arguments) != null ? _b : ""
								},
								hasFinished: false
							};
							const toolCall2 = toolCalls[index];
							if (((_c = toolCall2.function) == null ? void 0 : _c.name) != null && ((_d = toolCall2.function) == null ? void 0 : _d.arguments) != null) {
								if (toolCall2.function.arguments.length > 0) controller.enqueue({
									type: "tool-call-delta",
									toolCallType: "function",
									toolCallId: toolCall2.id,
									toolName: toolCall2.function.name,
									argsTextDelta: toolCall2.function.arguments
								});
								if (isParsableJson(toolCall2.function.arguments)) {
									controller.enqueue({
										type: "tool-call",
										toolCallType: "function",
										toolCallId: (_e = toolCall2.id) != null ? _e : generateId(),
										toolName: toolCall2.function.name,
										args: toolCall2.function.arguments
									});
									toolCall2.hasFinished = true;
								}
							}
							continue;
						}
						const toolCall = toolCalls[index];
						if (toolCall.hasFinished) continue;
						if (((_f = toolCallDelta.function) == null ? void 0 : _f.arguments) != null) toolCall.function.arguments += (_h = (_g = toolCallDelta.function) == null ? void 0 : _g.arguments) != null ? _h : "";
						controller.enqueue({
							type: "tool-call-delta",
							toolCallType: "function",
							toolCallId: toolCall.id,
							toolName: toolCall.function.name,
							argsTextDelta: (_i = toolCallDelta.function.arguments) != null ? _i : ""
						});
						if (((_j = toolCall.function) == null ? void 0 : _j.name) != null && ((_k = toolCall.function) == null ? void 0 : _k.arguments) != null && isParsableJson(toolCall.function.arguments)) {
							controller.enqueue({
								type: "tool-call",
								toolCallType: "function",
								toolCallId: (_l = toolCall.id) != null ? _l : generateId(),
								toolName: toolCall.function.name,
								args: toolCall.function.arguments
							});
							toolCall.hasFinished = true;
						}
					}
				},
				flush(controller) {
					var _a2, _b;
					const providerMetadata = {
						[providerOptionsName]: {},
						...metadataExtractor == null ? void 0 : metadataExtractor.buildMetadata()
					};
					if (usage.completionTokensDetails.reasoningTokens != null) providerMetadata[providerOptionsName].reasoningTokens = usage.completionTokensDetails.reasoningTokens;
					if (usage.completionTokensDetails.acceptedPredictionTokens != null) providerMetadata[providerOptionsName].acceptedPredictionTokens = usage.completionTokensDetails.acceptedPredictionTokens;
					if (usage.completionTokensDetails.rejectedPredictionTokens != null) providerMetadata[providerOptionsName].rejectedPredictionTokens = usage.completionTokensDetails.rejectedPredictionTokens;
					if (usage.promptTokensDetails.cachedTokens != null) providerMetadata[providerOptionsName].cachedPromptTokens = usage.promptTokensDetails.cachedTokens;
					controller.enqueue({
						type: "finish",
						finishReason,
						usage: {
							promptTokens: (_a2 = usage.promptTokens) != null ? _a2 : NaN,
							completionTokens: (_b = usage.completionTokens) != null ? _b : NaN
						},
						providerMetadata
					});
				}
			})),
			rawCall: {
				rawPrompt,
				rawSettings
			},
			rawResponse: { headers: responseHeaders },
			warnings,
			request: { body: JSON.stringify(body) }
		};
	}
};
var openaiCompatibleTokenUsageSchema = z.object({
	prompt_tokens: z.number().nullish(),
	completion_tokens: z.number().nullish(),
	prompt_tokens_details: z.object({ cached_tokens: z.number().nullish() }).nullish(),
	completion_tokens_details: z.object({
		reasoning_tokens: z.number().nullish(),
		accepted_prediction_tokens: z.number().nullish(),
		rejected_prediction_tokens: z.number().nullish()
	}).nullish()
}).nullish();
var OpenAICompatibleChatResponseSchema = z.object({
	id: z.string().nullish(),
	created: z.number().nullish(),
	model: z.string().nullish(),
	choices: z.array(z.object({
		message: z.object({
			role: z.literal("assistant").nullish(),
			content: z.string().nullish(),
			reasoning_content: z.string().nullish(),
			tool_calls: z.array(z.object({
				id: z.string().nullish(),
				type: z.literal("function"),
				function: z.object({
					name: z.string(),
					arguments: z.string()
				})
			})).nullish()
		}),
		finish_reason: z.string().nullish()
	})),
	usage: openaiCompatibleTokenUsageSchema
});
var createOpenAICompatibleChatChunkSchema = (errorSchema) => z.union([z.object({
	id: z.string().nullish(),
	created: z.number().nullish(),
	model: z.string().nullish(),
	choices: z.array(z.object({
		delta: z.object({
			role: z.enum(["assistant"]).nullish(),
			content: z.string().nullish(),
			reasoning_content: z.string().nullish(),
			tool_calls: z.array(z.object({
				index: z.number(),
				id: z.string().nullish(),
				type: z.literal("function").nullish(),
				function: z.object({
					name: z.string().nullish(),
					arguments: z.string().nullish()
				})
			})).nullish()
		}).nullish(),
		finish_reason: z.string().nullish()
	})),
	usage: openaiCompatibleTokenUsageSchema
}), errorSchema]);
function convertToOpenAICompatibleCompletionPrompt({ prompt, inputFormat, user = "user", assistant = "assistant" }) {
	if (inputFormat === "prompt" && prompt.length === 1 && prompt[0].role === "user" && prompt[0].content.length === 1 && prompt[0].content[0].type === "text") return { prompt: prompt[0].content[0].text };
	let text = "";
	if (prompt[0].role === "system") {
		text += `${prompt[0].content}

`;
		prompt = prompt.slice(1);
	}
	for (const { role, content } of prompt) switch (role) {
		case "system": throw new InvalidPromptError({
			message: "Unexpected system message in prompt: \${content}",
			prompt
		});
		case "user": {
			const userMessage = content.map((part) => {
				switch (part.type) {
					case "text": return part.text;
					case "image": throw new UnsupportedFunctionalityError({ functionality: "images" });
				}
			}).join("");
			text += `${user}:
${userMessage}

`;
			break;
		}
		case "assistant": {
			const assistantMessage = content.map((part) => {
				switch (part.type) {
					case "text": return part.text;
					case "tool-call": throw new UnsupportedFunctionalityError({ functionality: "tool-call messages" });
				}
			}).join("");
			text += `${assistant}:
${assistantMessage}

`;
			break;
		}
		case "tool": throw new UnsupportedFunctionalityError({ functionality: "tool messages" });
		default: {
			const _exhaustiveCheck = role;
			throw new Error(`Unsupported role: ${_exhaustiveCheck}`);
		}
	}
	text += `${assistant}:
`;
	return {
		prompt: text,
		stopSequences: [`
${user}:`]
	};
}
var OpenAICompatibleCompletionLanguageModel = class {
	constructor(modelId, settings, config) {
		this.specificationVersion = "v1";
		this.defaultObjectGenerationMode = void 0;
		var _a;
		this.modelId = modelId;
		this.settings = settings;
		this.config = config;
		const errorStructure = (_a = config.errorStructure) != null ? _a : defaultOpenAICompatibleErrorStructure;
		this.chunkSchema = createOpenAICompatibleCompletionChunkSchema(errorStructure.errorSchema);
		this.failedResponseHandler = createJsonErrorResponseHandler(errorStructure);
	}
	get provider() {
		return this.config.provider;
	}
	get providerOptionsName() {
		return this.config.provider.split(".")[0].trim();
	}
	getArgs({ mode, inputFormat, prompt, maxTokens, temperature, topP, topK, frequencyPenalty, presencePenalty, stopSequences: userStopSequences, responseFormat, seed, providerMetadata }) {
		var _a;
		const type = mode.type;
		const warnings = [];
		if (topK != null) warnings.push({
			type: "unsupported-setting",
			setting: "topK"
		});
		if (responseFormat != null && responseFormat.type !== "text") warnings.push({
			type: "unsupported-setting",
			setting: "responseFormat",
			details: "JSON response format is not supported."
		});
		const { prompt: completionPrompt, stopSequences } = convertToOpenAICompatibleCompletionPrompt({
			prompt,
			inputFormat
		});
		const stop = [...stopSequences != null ? stopSequences : [], ...userStopSequences != null ? userStopSequences : []];
		const baseArgs = {
			model: this.modelId,
			echo: this.settings.echo,
			logit_bias: this.settings.logitBias,
			suffix: this.settings.suffix,
			user: this.settings.user,
			max_tokens: maxTokens,
			temperature,
			top_p: topP,
			frequency_penalty: frequencyPenalty,
			presence_penalty: presencePenalty,
			seed,
			...providerMetadata == null ? void 0 : providerMetadata[this.providerOptionsName],
			prompt: completionPrompt,
			stop: stop.length > 0 ? stop : void 0
		};
		switch (type) {
			case "regular": {
				if ((_a = mode.tools) == null ? void 0 : _a.length) throw new UnsupportedFunctionalityError({ functionality: "tools" });
				if (mode.toolChoice) throw new UnsupportedFunctionalityError({ functionality: "toolChoice" });
				return {
					args: baseArgs,
					warnings
				};
			}
			case "object-json": throw new UnsupportedFunctionalityError({ functionality: "object-json mode" });
			case "object-tool": throw new UnsupportedFunctionalityError({ functionality: "object-tool mode" });
			default: {
				const _exhaustiveCheck = type;
				throw new Error(`Unsupported type: ${_exhaustiveCheck}`);
			}
		}
	}
	async doGenerate(options) {
		var _a, _b, _c, _d;
		const { args, warnings } = this.getArgs(options);
		const { responseHeaders, value: response, rawValue: rawResponse } = await postJsonToApi({
			url: this.config.url({
				path: "/completions",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), options.headers),
			body: args,
			failedResponseHandler: this.failedResponseHandler,
			successfulResponseHandler: createJsonResponseHandler(openaiCompatibleCompletionResponseSchema),
			abortSignal: options.abortSignal,
			fetch: this.config.fetch
		});
		const { prompt: rawPrompt,...rawSettings } = args;
		const choice = response.choices[0];
		return {
			text: choice.text,
			usage: {
				promptTokens: (_b = (_a = response.usage) == null ? void 0 : _a.prompt_tokens) != null ? _b : NaN,
				completionTokens: (_d = (_c = response.usage) == null ? void 0 : _c.completion_tokens) != null ? _d : NaN
			},
			finishReason: mapOpenAICompatibleFinishReason(choice.finish_reason),
			rawCall: {
				rawPrompt,
				rawSettings
			},
			rawResponse: {
				headers: responseHeaders,
				body: rawResponse
			},
			response: getResponseMetadata(response),
			warnings,
			request: { body: JSON.stringify(args) }
		};
	}
	async doStream(options) {
		const { args, warnings } = this.getArgs(options);
		const body = {
			...args,
			stream: true,
			stream_options: this.config.includeUsage ? { include_usage: true } : void 0
		};
		const { responseHeaders, value: response } = await postJsonToApi({
			url: this.config.url({
				path: "/completions",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), options.headers),
			body,
			failedResponseHandler: this.failedResponseHandler,
			successfulResponseHandler: createEventSourceResponseHandler(this.chunkSchema),
			abortSignal: options.abortSignal,
			fetch: this.config.fetch
		});
		const { prompt: rawPrompt,...rawSettings } = args;
		let finishReason = "unknown";
		let usage = {
			promptTokens: Number.NaN,
			completionTokens: Number.NaN
		};
		let isFirstChunk = true;
		return {
			stream: response.pipeThrough(new TransformStream({
				transform(chunk, controller) {
					if (!chunk.success) {
						finishReason = "error";
						controller.enqueue({
							type: "error",
							error: chunk.error
						});
						return;
					}
					const value = chunk.value;
					if ("error" in value) {
						finishReason = "error";
						controller.enqueue({
							type: "error",
							error: value.error
						});
						return;
					}
					if (isFirstChunk) {
						isFirstChunk = false;
						controller.enqueue({
							type: "response-metadata",
							...getResponseMetadata(value)
						});
					}
					if (value.usage != null) usage = {
						promptTokens: value.usage.prompt_tokens,
						completionTokens: value.usage.completion_tokens
					};
					const choice = value.choices[0];
					if ((choice == null ? void 0 : choice.finish_reason) != null) finishReason = mapOpenAICompatibleFinishReason(choice.finish_reason);
					if ((choice == null ? void 0 : choice.text) != null) controller.enqueue({
						type: "text-delta",
						textDelta: choice.text
					});
				},
				flush(controller) {
					controller.enqueue({
						type: "finish",
						finishReason,
						usage
					});
				}
			})),
			rawCall: {
				rawPrompt,
				rawSettings
			},
			rawResponse: { headers: responseHeaders },
			warnings,
			request: { body: JSON.stringify(body) }
		};
	}
};
var openaiCompatibleCompletionResponseSchema = z.object({
	id: z.string().nullish(),
	created: z.number().nullish(),
	model: z.string().nullish(),
	choices: z.array(z.object({
		text: z.string(),
		finish_reason: z.string()
	})),
	usage: z.object({
		prompt_tokens: z.number(),
		completion_tokens: z.number()
	}).nullish()
});
var createOpenAICompatibleCompletionChunkSchema = (errorSchema) => z.union([z.object({
	id: z.string().nullish(),
	created: z.number().nullish(),
	model: z.string().nullish(),
	choices: z.array(z.object({
		text: z.string(),
		finish_reason: z.string().nullish(),
		index: z.number()
	})),
	usage: z.object({
		prompt_tokens: z.number(),
		completion_tokens: z.number()
	}).nullish()
}), errorSchema]);
var OpenAICompatibleEmbeddingModel = class {
	constructor(modelId, settings, config) {
		this.specificationVersion = "v1";
		this.modelId = modelId;
		this.settings = settings;
		this.config = config;
	}
	get provider() {
		return this.config.provider;
	}
	get maxEmbeddingsPerCall() {
		var _a;
		return (_a = this.config.maxEmbeddingsPerCall) != null ? _a : 2048;
	}
	get supportsParallelCalls() {
		var _a;
		return (_a = this.config.supportsParallelCalls) != null ? _a : true;
	}
	async doEmbed({ values, headers, abortSignal }) {
		var _a;
		if (values.length > this.maxEmbeddingsPerCall) throw new TooManyEmbeddingValuesForCallError({
			provider: this.provider,
			modelId: this.modelId,
			maxEmbeddingsPerCall: this.maxEmbeddingsPerCall,
			values
		});
		const { responseHeaders, value: response } = await postJsonToApi({
			url: this.config.url({
				path: "/embeddings",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), headers),
			body: {
				model: this.modelId,
				input: values,
				encoding_format: "float",
				dimensions: this.settings.dimensions,
				user: this.settings.user
			},
			failedResponseHandler: createJsonErrorResponseHandler((_a = this.config.errorStructure) != null ? _a : defaultOpenAICompatibleErrorStructure),
			successfulResponseHandler: createJsonResponseHandler(openaiTextEmbeddingResponseSchema),
			abortSignal,
			fetch: this.config.fetch
		});
		return {
			embeddings: response.data.map((item) => item.embedding),
			usage: response.usage ? { tokens: response.usage.prompt_tokens } : void 0,
			rawResponse: { headers: responseHeaders }
		};
	}
};
var openaiTextEmbeddingResponseSchema = z.object({
	data: z.array(z.object({ embedding: z.array(z.number()) })),
	usage: z.object({ prompt_tokens: z.number() }).nullish()
});
var OpenAICompatibleImageModel = class {
	constructor(modelId, settings, config) {
		this.modelId = modelId;
		this.settings = settings;
		this.config = config;
		this.specificationVersion = "v1";
	}
	get maxImagesPerCall() {
		var _a;
		return (_a = this.settings.maxImagesPerCall) != null ? _a : 10;
	}
	get provider() {
		return this.config.provider;
	}
	async doGenerate({ prompt, n, size, aspectRatio, seed, providerOptions, headers, abortSignal }) {
		var _a, _b, _c, _d, _e;
		const warnings = [];
		if (aspectRatio != null) warnings.push({
			type: "unsupported-setting",
			setting: "aspectRatio",
			details: "This model does not support aspect ratio. Use `size` instead."
		});
		if (seed != null) warnings.push({
			type: "unsupported-setting",
			setting: "seed"
		});
		const currentDate = (_c = (_b = (_a = this.config._internal) == null ? void 0 : _a.currentDate) == null ? void 0 : _b.call(_a)) != null ? _c : /* @__PURE__ */ new Date();
		const { value: response, responseHeaders } = await postJsonToApi({
			url: this.config.url({
				path: "/images/generations",
				modelId: this.modelId
			}),
			headers: combineHeaders(this.config.headers(), headers),
			body: {
				model: this.modelId,
				prompt,
				n,
				size,
				...(_d = providerOptions.openai) != null ? _d : {},
				response_format: "b64_json",
				...this.settings.user ? { user: this.settings.user } : {}
			},
			failedResponseHandler: createJsonErrorResponseHandler((_e = this.config.errorStructure) != null ? _e : defaultOpenAICompatibleErrorStructure),
			successfulResponseHandler: createJsonResponseHandler(openaiCompatibleImageResponseSchema),
			abortSignal,
			fetch: this.config.fetch
		});
		return {
			images: response.data.map((item) => item.b64_json),
			warnings,
			response: {
				timestamp: currentDate,
				modelId: this.modelId,
				headers: responseHeaders
			}
		};
	}
};
var openaiCompatibleImageResponseSchema = z.object({ data: z.array(z.object({ b64_json: z.string() })) });
function createOpenAICompatible(options) {
	const baseURL = withoutTrailingSlash(options.baseURL);
	const providerName = options.name;
	const getHeaders = () => ({
		...options.apiKey && { Authorization: `Bearer ${options.apiKey}` },
		...options.headers
	});
	const getCommonModelConfig = (modelType) => ({
		provider: `${providerName}.${modelType}`,
		url: ({ path: path$1 }) => {
			const url = new URL(`${baseURL}${path$1}`);
			if (options.queryParams) url.search = new URLSearchParams(options.queryParams).toString();
			return url.toString();
		},
		headers: getHeaders,
		fetch: options.fetch
	});
	const createLanguageModel = (modelId, settings = {}) => createChatModel(modelId, settings);
	const createChatModel = (modelId, settings = {}) => new OpenAICompatibleChatLanguageModel(modelId, settings, {
		...getCommonModelConfig("chat"),
		defaultObjectGenerationMode: "tool"
	});
	const createCompletionModel = (modelId, settings = {}) => new OpenAICompatibleCompletionLanguageModel(modelId, settings, getCommonModelConfig("completion"));
	const createEmbeddingModel = (modelId, settings = {}) => new OpenAICompatibleEmbeddingModel(modelId, settings, getCommonModelConfig("embedding"));
	const createImageModel = (modelId, settings = {}) => new OpenAICompatibleImageModel(modelId, settings, getCommonModelConfig("image"));
	const provider = (modelId, settings) => createLanguageModel(modelId, settings);
	provider.languageModel = createLanguageModel;
	provider.chatModel = createChatModel;
	provider.completionModel = createCompletionModel;
	provider.textEmbeddingModel = createEmbeddingModel;
	provider.imageModel = createImageModel;
	return provider;
}

//#endregion
export { createOpenAICompatible, createStorage, fs_lite_default, generateId, safeParseJSON };