import { __commonJS, __require, __toESM } from "./chunk-BlwiYZwt.js";
import { mkdir } from "node:fs/promises";
import fs$1, { promises } from "node:fs";

//#region node_modules/cachedir/index.js
var require_cachedir = __commonJS({ "node_modules/cachedir/index.js"(exports, module) {
	const os = __require("os");
	const path = __require("path");
	function posix(id) {
		const cacheHome = process.env.XDG_CACHE_HOME || path.join(os.homedir(), ".cache");
		return path.join(cacheHome, id);
	}
	function darwin(id) {
		return path.join(os.homedir(), "Library", "Caches", id);
	}
	function win32(id) {
		const appData = process.env.LOCALAPPDATA || path.join(os.homedir(), "AppData", "Local");
		return path.join(appData, id, "Cache");
	}
	const implementation = function() {
		switch (os.platform()) {
			case "darwin": return darwin;
			case "win32": return win32;
			case "aix":
			case "android":
			case "freebsd":
			case "linux":
			case "netbsd":
			case "openbsd":
			case "sunos": return posix;
			default:
				console.error(`(node:${process.pid}) [cachedir] Warning: the platform "${os.platform()}" is not currently supported by node-cachedir, falling back to "posix". Please file an issue with your platform here: https://github.com/LinusU/node-cachedir/issues/new`);
				return posix;
		}
	}();
	module.exports = function cachedir(id) {
		if (typeof id !== "string") throw new TypeError("id is not a string");
		if (id.length === 0) throw new Error("id cannot be empty");
		if (/[^0-9a-zA-Z-]/.test(id)) throw new Error("id cannot contain special characters");
		return implementation(id);
	};
} });
var import_cachedir = __toESM(require_cachedir());

//#endregion
//#region node_modules/path-exists/index.js
async function pathExists(path$1) {
	try {
		await promises.access(path$1);
		return true;
	} catch {
		return false;
	}
}

//#endregion
//#region node_modules/global-cache-dir/build/src/main.js
const globalCacheDir = async (name) => {
	const cacheDir = (0, import_cachedir.default)(name);
	await createCacheDir(cacheDir);
	return cacheDir;
};
var main_default = globalCacheDir;
const createCacheDir = async (cacheDir) => {
	if (await pathExists(cacheDir)) return;
	await mkdir(cacheDir, { recursive: true });
};

//#endregion
export { main_default };