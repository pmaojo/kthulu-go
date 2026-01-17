import { main_default } from "./assets/main-Cbt18fNR.js";
import { createOpenAICompatible, createStorage, fs_lite_default } from "./assets/dist-CwZ8V55S.js";
import { import_react, z } from "./assets/react-C-QzwZjq.js";
import path from "node:path";
import { anthropic, createAnthropic } from "@ai-sdk/anthropic";
import { createGoogleGenerativeAI, google } from "@ai-sdk/google";
import { createOpenAI, openai } from "@ai-sdk/openai";

//#region src/lib.ts
const cacheDir = await main_default("OpenCoder");
const storage = createStorage({ driver: fs_lite_default({ base: path.join(cacheDir, "general-cache") }) });

//#endregion
var React = import_react.default;
export { React, anthropic, createAnthropic, createGoogleGenerativeAI, createOpenAI, createOpenAICompatible, google, openai, storage, z };