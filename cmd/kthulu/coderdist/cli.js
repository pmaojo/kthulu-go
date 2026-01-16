#!/usr/bin/env node
import { main_default } from "./assets/main-Cbt18fNR.js";
import "./assets/source-map-support-BDjyShzz.js";
import "./assets/register-DzYslTyX.js";
import module from "node:module";
import path from "node:path";

//#region src/index.ts
try {
	const cacheDir = await main_default("OpenCoder");
	module.enableCompileCache(path.join(cacheDir, "v8-cache"));
	if (process.argv.includes("--empty-cache")) module.flushCompileCache();
} catch {}
globalThis.window = {};
globalThis.stop = () => {};
import("./assets/main-CP2UZyro.js");

//#endregion