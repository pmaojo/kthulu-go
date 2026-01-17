import { __commonJS, __esm, __export, __require, __toCommonJS, __toESM } from "./chunk-BlwiYZwt.js";
import "./source-map-support-BDjyShzz.js";
import "./register-DzYslTyX.js";
import "./dist-CwZ8V55S.js";
import "./react-C-QzwZjq.js";
import { dedent_default } from "./mcp-DX8bwdY8.js";
import { config, createCoder, detect, env, mod_default, package_default, source_default, stringWidth, widestLine, wrapAnsi } from "./core-CgUgMPqR.js";
import { appendFileSync, existsSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import process$1 from "node:process";
import cliBoxes from "cli-boxes";

//#region node_modules/camelcase/index.js
const UPPERCASE = /[\p{Lu}]/u;
const LOWERCASE = /[\p{Ll}]/u;
const LEADING_CAPITAL = /^[\p{Lu}](?![\p{Lu}])/gu;
const IDENTIFIER = /([\p{Alpha}\p{N}_]|$)/u;
const SEPARATORS = /[_.\- ]+/;
const LEADING_SEPARATORS = new RegExp("^" + SEPARATORS.source);
const SEPARATORS_AND_IDENTIFIER = new RegExp(SEPARATORS.source + IDENTIFIER.source, "gu");
const NUMBERS_AND_IDENTIFIER = new RegExp("\\d+" + IDENTIFIER.source, "gu");
const preserveCamelCase = (string, toLowerCase, toUpperCase, preserveConsecutiveUppercase$1) => {
	let isLastCharLower = false;
	let isLastCharUpper = false;
	let isLastLastCharUpper = false;
	let isLastLastCharPreserved = false;
	for (let index = 0; index < string.length; index++) {
		const character = string[index];
		isLastLastCharPreserved = index > 2 ? string[index - 3] === "-" : true;
		if (isLastCharLower && UPPERCASE.test(character)) {
			string = string.slice(0, index) + "-" + string.slice(index);
			isLastCharLower = false;
			isLastLastCharUpper = isLastCharUpper;
			isLastCharUpper = true;
			index++;
		} else if (isLastCharUpper && isLastLastCharUpper && LOWERCASE.test(character) && (!isLastLastCharPreserved || preserveConsecutiveUppercase$1)) {
			string = string.slice(0, index - 1) + "-" + string.slice(index - 1);
			isLastLastCharUpper = isLastCharUpper;
			isLastCharUpper = false;
			isLastCharLower = true;
		} else {
			isLastCharLower = toLowerCase(character) === character && toUpperCase(character) !== character;
			isLastLastCharUpper = isLastCharUpper;
			isLastCharUpper = toUpperCase(character) === character && toLowerCase(character) !== character;
		}
	}
	return string;
};
const preserveConsecutiveUppercase = (input, toLowerCase) => {
	LEADING_CAPITAL.lastIndex = 0;
	return input.replaceAll(LEADING_CAPITAL, (match) => toLowerCase(match));
};
const postProcess = (input, toUpperCase) => {
	SEPARATORS_AND_IDENTIFIER.lastIndex = 0;
	NUMBERS_AND_IDENTIFIER.lastIndex = 0;
	return input.replaceAll(NUMBERS_AND_IDENTIFIER, (match, pattern, offset) => ["_", "-"].includes(input.charAt(offset + match.length)) ? match : toUpperCase(match)).replaceAll(SEPARATORS_AND_IDENTIFIER, (_, identifier) => toUpperCase(identifier));
};
function camelCase(input, options) {
	if (!(typeof input === "string" || Array.isArray(input))) throw new TypeError("Expected the input to be `string | string[]`");
	options = {
		pascalCase: false,
		preserveConsecutiveUppercase: false,
		...options
	};
	if (Array.isArray(input)) input = input.map((x) => x.trim()).filter((x) => x.length).join("-");
	else input = input.trim();
	if (input.length === 0) return "";
	const toLowerCase = options.locale === false ? (string) => string.toLowerCase() : (string) => string.toLocaleLowerCase(options.locale);
	const toUpperCase = options.locale === false ? (string) => string.toUpperCase() : (string) => string.toLocaleUpperCase(options.locale);
	if (input.length === 1) {
		if (SEPARATORS.test(input)) return "";
		return options.pascalCase ? toUpperCase(input) : toLowerCase(input);
	}
	const hasUpperCase = input !== toLowerCase(input);
	if (hasUpperCase) input = preserveCamelCase(input, toLowerCase, toUpperCase, options.preserveConsecutiveUppercase);
	input = input.replace(LEADING_SEPARATORS, "");
	input = options.preserveConsecutiveUppercase ? preserveConsecutiveUppercase(input, toLowerCase) : toLowerCase(input);
	if (options.pascalCase) input = toUpperCase(input.charAt(0)) + input.slice(1);
	return postProcess(input, toUpperCase);
}

//#endregion
//#region node_modules/ansi-align/node_modules/string-width/node_modules/strip-ansi/node_modules/ansi-regex/index.js
var require_ansi_regex = __commonJS({ "node_modules/ansi-align/node_modules/string-width/node_modules/strip-ansi/node_modules/ansi-regex/index.js"(exports, module) {
	module.exports = ({ onlyFirst = false } = {}) => {
		const pattern = ["[\\u001B\\u009B][[\\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]+)*|[a-zA-Z\\d]+(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\\u0007)", "(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"].join("|");
		return new RegExp(pattern, onlyFirst ? void 0 : "g");
	};
} });

//#endregion
//#region node_modules/ansi-align/node_modules/string-width/node_modules/strip-ansi/index.js
var require_strip_ansi = __commonJS({ "node_modules/ansi-align/node_modules/string-width/node_modules/strip-ansi/index.js"(exports, module) {
	const ansiRegex = require_ansi_regex();
	module.exports = (string) => typeof string === "string" ? string.replace(ansiRegex(), "") : string;
} });

//#endregion
//#region node_modules/ansi-align/node_modules/string-width/node_modules/is-fullwidth-code-point/index.js
var require_is_fullwidth_code_point = __commonJS({ "node_modules/ansi-align/node_modules/string-width/node_modules/is-fullwidth-code-point/index.js"(exports, module) {
	const isFullwidthCodePoint$1 = (codePoint) => {
		if (Number.isNaN(codePoint)) return false;
		if (codePoint >= 4352 && (codePoint <= 4447 || codePoint === 9001 || codePoint === 9002 || 11904 <= codePoint && codePoint <= 12871 && codePoint !== 12351 || 12880 <= codePoint && codePoint <= 19903 || 19968 <= codePoint && codePoint <= 42182 || 43360 <= codePoint && codePoint <= 43388 || 44032 <= codePoint && codePoint <= 55203 || 63744 <= codePoint && codePoint <= 64255 || 65040 <= codePoint && codePoint <= 65049 || 65072 <= codePoint && codePoint <= 65131 || 65281 <= codePoint && codePoint <= 65376 || 65504 <= codePoint && codePoint <= 65510 || 110592 <= codePoint && codePoint <= 110593 || 127488 <= codePoint && codePoint <= 127569 || 131072 <= codePoint && codePoint <= 262141)) return true;
		return false;
	};
	module.exports = isFullwidthCodePoint$1;
	module.exports.default = isFullwidthCodePoint$1;
} });

//#endregion
//#region node_modules/ansi-align/node_modules/string-width/node_modules/emoji-regex/index.js
var require_emoji_regex = __commonJS({ "node_modules/ansi-align/node_modules/string-width/node_modules/emoji-regex/index.js"(exports, module) {
	module.exports = function() {
		return /\uD83C\uDFF4\uDB40\uDC67\uDB40\uDC62(?:\uDB40\uDC65\uDB40\uDC6E\uDB40\uDC67|\uDB40\uDC73\uDB40\uDC63\uDB40\uDC74|\uDB40\uDC77\uDB40\uDC6C\uDB40\uDC73)\uDB40\uDC7F|\uD83D\uDC68(?:\uD83C\uDFFC\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68\uD83C\uDFFB|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFF\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFE])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFE\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFD])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFD\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB\uDFFC])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\u200D(?:\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83D\uDC68|(?:\uD83D[\uDC68\uDC69])\u200D(?:\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67]))|\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67])|(?:\uD83D[\uDC68\uDC69])\u200D(?:\uD83D[\uDC66\uDC67])|[\u2695\u2696\u2708]\uFE0F|\uD83D[\uDC66\uDC67]|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|(?:\uD83C\uDFFB\u200D[\u2695\u2696\u2708]|\uD83C\uDFFF\u200D[\u2695\u2696\u2708]|\uD83C\uDFFE\u200D[\u2695\u2696\u2708]|\uD83C\uDFFD\u200D[\u2695\u2696\u2708]|\uD83C\uDFFC\u200D[\u2695\u2696\u2708])\uFE0F|\uD83C\uDFFB\u200D(?:\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C[\uDFFB-\uDFFF])|(?:\uD83E\uDDD1\uD83C\uDFFB\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFC\u200D\uD83E\uDD1D\u200D\uD83D\uDC69)\uD83C\uDFFB|\uD83E\uDDD1(?:\uD83C\uDFFF\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1(?:\uD83C[\uDFFB-\uDFFF])|\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1)|(?:\uD83E\uDDD1\uD83C\uDFFE\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFF\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFB-\uDFFE])|(?:\uD83E\uDDD1\uD83C\uDFFC\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFD\u200D\uD83E\uDD1D\u200D\uD83D\uDC69)(?:\uD83C[\uDFFB\uDFFC])|\uD83D\uDC69(?:\uD83C\uDFFE\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFD\uDFFF])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFC\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB\uDFFD-\uDFFF])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFB\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFC-\uDFFF])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFD\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB\uDFFC\uDFFE\uDFFF])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\u200D(?:\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D(?:\uD83D[\uDC68\uDC69])|\uD83D[\uDC68\uDC69])|\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFF\u200D(?:\uD83C[\uDF3E\uDF73\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD]))|\uD83D\uDC69\u200D\uD83D\uDC69\u200D(?:\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67]))|(?:\uD83E\uDDD1\uD83C\uDFFD\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFE\u200D\uD83E\uDD1D\u200D\uD83D\uDC69)(?:\uD83C[\uDFFB-\uDFFD])|\uD83D\uDC69\u200D\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC69\u200D\uD83D\uDC69\u200D(?:\uD83D[\uDC66\uDC67])|(?:\uD83D\uDC41\uFE0F\u200D\uD83D\uDDE8|\uD83D\uDC69(?:\uD83C\uDFFF\u200D[\u2695\u2696\u2708]|\uD83C\uDFFE\u200D[\u2695\u2696\u2708]|\uD83C\uDFFC\u200D[\u2695\u2696\u2708]|\uD83C\uDFFB\u200D[\u2695\u2696\u2708]|\uD83C\uDFFD\u200D[\u2695\u2696\u2708]|\u200D[\u2695\u2696\u2708])|(?:(?:\u26F9|\uD83C[\uDFCB\uDFCC]|\uD83D\uDD75)\uFE0F|\uD83D\uDC6F|\uD83E[\uDD3C\uDDDE\uDDDF])\u200D[\u2640\u2642]|(?:\u26F9|\uD83C[\uDFCB\uDFCC]|\uD83D\uDD75)(?:\uD83C[\uDFFB-\uDFFF])\u200D[\u2640\u2642]|(?:\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD6-\uDDDD])(?:(?:\uD83C[\uDFFB-\uDFFF])\u200D[\u2640\u2642]|\u200D[\u2640\u2642])|\uD83C\uDFF4\u200D\u2620)\uFE0F|\uD83D\uDC69\u200D\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67])|\uD83C\uDFF3\uFE0F\u200D\uD83C\uDF08|\uD83D\uDC15\u200D\uD83E\uDDBA|\uD83D\uDC69\u200D\uD83D\uDC66|\uD83D\uDC69\u200D\uD83D\uDC67|\uD83C\uDDFD\uD83C\uDDF0|\uD83C\uDDF4\uD83C\uDDF2|\uD83C\uDDF6\uD83C\uDDE6|[#\*0-9]\uFE0F\u20E3|\uD83C\uDDE7(?:\uD83C[\uDDE6\uDDE7\uDDE9-\uDDEF\uDDF1-\uDDF4\uDDF6-\uDDF9\uDDFB\uDDFC\uDDFE\uDDFF])|\uD83C\uDDF9(?:\uD83C[\uDDE6\uDDE8\uDDE9\uDDEB-\uDDED\uDDEF-\uDDF4\uDDF7\uDDF9\uDDFB\uDDFC\uDDFF])|\uD83C\uDDEA(?:\uD83C[\uDDE6\uDDE8\uDDEA\uDDEC\uDDED\uDDF7-\uDDFA])|\uD83E\uDDD1(?:\uD83C[\uDFFB-\uDFFF])|\uD83C\uDDF7(?:\uD83C[\uDDEA\uDDF4\uDDF8\uDDFA\uDDFC])|\uD83D\uDC69(?:\uD83C[\uDFFB-\uDFFF])|\uD83C\uDDF2(?:\uD83C[\uDDE6\uDDE8-\uDDED\uDDF0-\uDDFF])|\uD83C\uDDE6(?:\uD83C[\uDDE8-\uDDEC\uDDEE\uDDF1\uDDF2\uDDF4\uDDF6-\uDDFA\uDDFC\uDDFD\uDDFF])|\uD83C\uDDF0(?:\uD83C[\uDDEA\uDDEC-\uDDEE\uDDF2\uDDF3\uDDF5\uDDF7\uDDFC\uDDFE\uDDFF])|\uD83C\uDDED(?:\uD83C[\uDDF0\uDDF2\uDDF3\uDDF7\uDDF9\uDDFA])|\uD83C\uDDE9(?:\uD83C[\uDDEA\uDDEC\uDDEF\uDDF0\uDDF2\uDDF4\uDDFF])|\uD83C\uDDFE(?:\uD83C[\uDDEA\uDDF9])|\uD83C\uDDEC(?:\uD83C[\uDDE6\uDDE7\uDDE9-\uDDEE\uDDF1-\uDDF3\uDDF5-\uDDFA\uDDFC\uDDFE])|\uD83C\uDDF8(?:\uD83C[\uDDE6-\uDDEA\uDDEC-\uDDF4\uDDF7-\uDDF9\uDDFB\uDDFD-\uDDFF])|\uD83C\uDDEB(?:\uD83C[\uDDEE-\uDDF0\uDDF2\uDDF4\uDDF7])|\uD83C\uDDF5(?:\uD83C[\uDDE6\uDDEA-\uDDED\uDDF0-\uDDF3\uDDF7-\uDDF9\uDDFC\uDDFE])|\uD83C\uDDFB(?:\uD83C[\uDDE6\uDDE8\uDDEA\uDDEC\uDDEE\uDDF3\uDDFA])|\uD83C\uDDF3(?:\uD83C[\uDDE6\uDDE8\uDDEA-\uDDEC\uDDEE\uDDF1\uDDF4\uDDF5\uDDF7\uDDFA\uDDFF])|\uD83C\uDDE8(?:\uD83C[\uDDE6\uDDE8\uDDE9\uDDEB-\uDDEE\uDDF0-\uDDF5\uDDF7\uDDFA-\uDDFF])|\uD83C\uDDF1(?:\uD83C[\uDDE6-\uDDE8\uDDEE\uDDF0\uDDF7-\uDDFB\uDDFE])|\uD83C\uDDFF(?:\uD83C[\uDDE6\uDDF2\uDDFC])|\uD83C\uDDFC(?:\uD83C[\uDDEB\uDDF8])|\uD83C\uDDFA(?:\uD83C[\uDDE6\uDDEC\uDDF2\uDDF3\uDDF8\uDDFE\uDDFF])|\uD83C\uDDEE(?:\uD83C[\uDDE8-\uDDEA\uDDF1-\uDDF4\uDDF6-\uDDF9])|\uD83C\uDDEF(?:\uD83C[\uDDEA\uDDF2\uDDF4\uDDF5])|(?:\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD6-\uDDDD])(?:\uD83C[\uDFFB-\uDFFF])|(?:\u26F9|\uD83C[\uDFCB\uDFCC]|\uD83D\uDD75)(?:\uD83C[\uDFFB-\uDFFF])|(?:[\u261D\u270A-\u270D]|\uD83C[\uDF85\uDFC2\uDFC7]|\uD83D[\uDC42\uDC43\uDC46-\uDC50\uDC66\uDC67\uDC6B-\uDC6D\uDC70\uDC72\uDC74-\uDC76\uDC78\uDC7C\uDC83\uDC85\uDCAA\uDD74\uDD7A\uDD90\uDD95\uDD96\uDE4C\uDE4F\uDEC0\uDECC]|\uD83E[\uDD0F\uDD18-\uDD1C\uDD1E\uDD1F\uDD30-\uDD36\uDDB5\uDDB6\uDDBB\uDDD2-\uDDD5])(?:\uD83C[\uDFFB-\uDFFF])|(?:[\u231A\u231B\u23E9-\u23EC\u23F0\u23F3\u25FD\u25FE\u2614\u2615\u2648-\u2653\u267F\u2693\u26A1\u26AA\u26AB\u26BD\u26BE\u26C4\u26C5\u26CE\u26D4\u26EA\u26F2\u26F3\u26F5\u26FA\u26FD\u2705\u270A\u270B\u2728\u274C\u274E\u2753-\u2755\u2757\u2795-\u2797\u27B0\u27BF\u2B1B\u2B1C\u2B50\u2B55]|\uD83C[\uDC04\uDCCF\uDD8E\uDD91-\uDD9A\uDDE6-\uDDFF\uDE01\uDE1A\uDE2F\uDE32-\uDE36\uDE38-\uDE3A\uDE50\uDE51\uDF00-\uDF20\uDF2D-\uDF35\uDF37-\uDF7C\uDF7E-\uDF93\uDFA0-\uDFCA\uDFCF-\uDFD3\uDFE0-\uDFF0\uDFF4\uDFF8-\uDFFF]|\uD83D[\uDC00-\uDC3E\uDC40\uDC42-\uDCFC\uDCFF-\uDD3D\uDD4B-\uDD4E\uDD50-\uDD67\uDD7A\uDD95\uDD96\uDDA4\uDDFB-\uDE4F\uDE80-\uDEC5\uDECC\uDED0-\uDED2\uDED5\uDEEB\uDEEC\uDEF4-\uDEFA\uDFE0-\uDFEB]|\uD83E[\uDD0D-\uDD3A\uDD3C-\uDD45\uDD47-\uDD71\uDD73-\uDD76\uDD7A-\uDDA2\uDDA5-\uDDAA\uDDAE-\uDDCA\uDDCD-\uDDFF\uDE70-\uDE73\uDE78-\uDE7A\uDE80-\uDE82\uDE90-\uDE95])|(?:[#\*0-9\xA9\xAE\u203C\u2049\u2122\u2139\u2194-\u2199\u21A9\u21AA\u231A\u231B\u2328\u23CF\u23E9-\u23F3\u23F8-\u23FA\u24C2\u25AA\u25AB\u25B6\u25C0\u25FB-\u25FE\u2600-\u2604\u260E\u2611\u2614\u2615\u2618\u261D\u2620\u2622\u2623\u2626\u262A\u262E\u262F\u2638-\u263A\u2640\u2642\u2648-\u2653\u265F\u2660\u2663\u2665\u2666\u2668\u267B\u267E\u267F\u2692-\u2697\u2699\u269B\u269C\u26A0\u26A1\u26AA\u26AB\u26B0\u26B1\u26BD\u26BE\u26C4\u26C5\u26C8\u26CE\u26CF\u26D1\u26D3\u26D4\u26E9\u26EA\u26F0-\u26F5\u26F7-\u26FA\u26FD\u2702\u2705\u2708-\u270D\u270F\u2712\u2714\u2716\u271D\u2721\u2728\u2733\u2734\u2744\u2747\u274C\u274E\u2753-\u2755\u2757\u2763\u2764\u2795-\u2797\u27A1\u27B0\u27BF\u2934\u2935\u2B05-\u2B07\u2B1B\u2B1C\u2B50\u2B55\u3030\u303D\u3297\u3299]|\uD83C[\uDC04\uDCCF\uDD70\uDD71\uDD7E\uDD7F\uDD8E\uDD91-\uDD9A\uDDE6-\uDDFF\uDE01\uDE02\uDE1A\uDE2F\uDE32-\uDE3A\uDE50\uDE51\uDF00-\uDF21\uDF24-\uDF93\uDF96\uDF97\uDF99-\uDF9B\uDF9E-\uDFF0\uDFF3-\uDFF5\uDFF7-\uDFFF]|\uD83D[\uDC00-\uDCFD\uDCFF-\uDD3D\uDD49-\uDD4E\uDD50-\uDD67\uDD6F\uDD70\uDD73-\uDD7A\uDD87\uDD8A-\uDD8D\uDD90\uDD95\uDD96\uDDA4\uDDA5\uDDA8\uDDB1\uDDB2\uDDBC\uDDC2-\uDDC4\uDDD1-\uDDD3\uDDDC-\uDDDE\uDDE1\uDDE3\uDDE8\uDDEF\uDDF3\uDDFA-\uDE4F\uDE80-\uDEC5\uDECB-\uDED2\uDED5\uDEE0-\uDEE5\uDEE9\uDEEB\uDEEC\uDEF0\uDEF3-\uDEFA\uDFE0-\uDFEB]|\uD83E[\uDD0D-\uDD3A\uDD3C-\uDD45\uDD47-\uDD71\uDD73-\uDD76\uDD7A-\uDDA2\uDDA5-\uDDAA\uDDAE-\uDDCA\uDDCD-\uDDFF\uDE70-\uDE73\uDE78-\uDE7A\uDE80-\uDE82\uDE90-\uDE95])\uFE0F|(?:[\u261D\u26F9\u270A-\u270D]|\uD83C[\uDF85\uDFC2-\uDFC4\uDFC7\uDFCA-\uDFCC]|\uD83D[\uDC42\uDC43\uDC46-\uDC50\uDC66-\uDC78\uDC7C\uDC81-\uDC83\uDC85-\uDC87\uDC8F\uDC91\uDCAA\uDD74\uDD75\uDD7A\uDD90\uDD95\uDD96\uDE45-\uDE47\uDE4B-\uDE4F\uDEA3\uDEB4-\uDEB6\uDEC0\uDECC]|\uD83E[\uDD0F\uDD18-\uDD1F\uDD26\uDD30-\uDD39\uDD3C-\uDD3E\uDDB5\uDDB6\uDDB8\uDDB9\uDDBB\uDDCD-\uDDCF\uDDD1-\uDDDD])/g;
	};
} });

//#endregion
//#region node_modules/ansi-align/node_modules/string-width/index.js
var require_string_width = __commonJS({ "node_modules/ansi-align/node_modules/string-width/index.js"(exports, module) {
	const stripAnsi = require_strip_ansi();
	const isFullwidthCodePoint = require_is_fullwidth_code_point();
	const emojiRegex = require_emoji_regex();
	const stringWidth$2 = (string) => {
		if (typeof string !== "string" || string.length === 0) return 0;
		string = stripAnsi(string);
		if (string.length === 0) return 0;
		string = string.replace(emojiRegex(), "  ");
		let width = 0;
		for (let i = 0; i < string.length; i++) {
			const code = string.codePointAt(i);
			if (code <= 31 || code >= 127 && code <= 159) continue;
			if (code >= 768 && code <= 879) continue;
			if (code > 65535) i++;
			width += isFullwidthCodePoint(code) ? 2 : 1;
		}
		return width;
	};
	module.exports = stringWidth$2;
	module.exports.default = stringWidth$2;
} });

//#endregion
//#region node_modules/ansi-align/index.js
var require_ansi_align = __commonJS({ "node_modules/ansi-align/index.js"(exports, module) {
	const stringWidth$1 = require_string_width();
	function ansiAlign(text, opts) {
		if (!text) return text;
		opts = opts || {};
		const align = opts.align || "center";
		if (align === "left") return text;
		const split = opts.split || "\n";
		const pad = opts.pad || " ";
		const widthDiffFn = align !== "right" ? halfDiff : fullDiff;
		let returnString = false;
		if (!Array.isArray(text)) {
			returnString = true;
			text = String(text).split(split);
		}
		let width;
		let maxWidth = 0;
		text = text.map(function(str) {
			str = String(str);
			width = stringWidth$1(str);
			maxWidth = Math.max(width, maxWidth);
			return {
				str,
				width
			};
		}).map(function(obj) {
			return new Array(widthDiffFn(maxWidth, obj.width) + 1).join(pad) + obj.str;
		});
		return returnString ? text.join(split) : text;
	}
	ansiAlign.left = function left(text) {
		return ansiAlign(text, { align: "left" });
	};
	ansiAlign.center = function center(text) {
		return ansiAlign(text, { align: "center" });
	};
	ansiAlign.right = function right(text) {
		return ansiAlign(text, { align: "right" });
	};
	module.exports = ansiAlign;
	function halfDiff(maxWidth, curWidth) {
		return Math.floor((maxWidth - curWidth) / 2);
	}
	function fullDiff(maxWidth, curWidth) {
		return maxWidth - curWidth;
	}
} });
var import_ansi_align = __toESM(require_ansi_align());

//#endregion
//#region node_modules/boxen/index.js
const NEWLINE = "\n";
const PAD = " ";
const NONE = "none";
const terminalColumns = () => {
	const { env: env$2, stdout, stderr } = process$1;
	if (stdout?.columns) return stdout.columns;
	if (stderr?.columns) return stderr.columns;
	if (env$2.COLUMNS) return Number.parseInt(env$2.COLUMNS, 10);
	return 80;
};
const getObject = (detail) => typeof detail === "number" ? {
	top: detail,
	right: detail * 3,
	bottom: detail,
	left: detail * 3
} : {
	top: 0,
	right: 0,
	bottom: 0,
	left: 0,
	...detail
};
const getBorderWidth = (borderStyle) => borderStyle === NONE ? 0 : 2;
const getBorderChars = (borderStyle) => {
	const sides = [
		"topLeft",
		"topRight",
		"bottomRight",
		"bottomLeft",
		"left",
		"right",
		"top",
		"bottom"
	];
	let characters;
	if (borderStyle === NONE) {
		borderStyle = {};
		for (const side of sides) borderStyle[side] = "";
	}
	if (typeof borderStyle === "string") {
		characters = cliBoxes[borderStyle];
		if (!characters) throw new TypeError(`Invalid border style: ${borderStyle}`);
	} else {
		if (typeof borderStyle?.vertical === "string") {
			borderStyle.left = borderStyle.vertical;
			borderStyle.right = borderStyle.vertical;
		}
		if (typeof borderStyle?.horizontal === "string") {
			borderStyle.top = borderStyle.horizontal;
			borderStyle.bottom = borderStyle.horizontal;
		}
		for (const side of sides) if (borderStyle[side] === null || typeof borderStyle[side] !== "string") throw new TypeError(`Invalid border style: ${side}`);
		characters = borderStyle;
	}
	return characters;
};
const makeTitle = (text, horizontal, alignment) => {
	let title = "";
	const textWidth = stringWidth(text);
	switch (alignment) {
		case "left": {
			title = text + horizontal.slice(textWidth);
			break;
		}
		case "right": {
			title = horizontal.slice(textWidth) + text;
			break;
		}
		default: {
			horizontal = horizontal.slice(textWidth);
			if (horizontal.length % 2 === 1) {
				horizontal = horizontal.slice(Math.floor(horizontal.length / 2));
				title = horizontal.slice(1) + text + horizontal;
			} else {
				horizontal = horizontal.slice(horizontal.length / 2);
				title = horizontal + text + horizontal;
			}
			break;
		}
	}
	return title;
};
const makeContentText = (text, { padding, width, textAlignment, height }) => {
	text = (0, import_ansi_align.default)(text, { align: textAlignment });
	let lines = text.split(NEWLINE);
	const textWidth = widestLine(text);
	const max = width - padding.left - padding.right;
	if (textWidth > max) {
		const newLines = [];
		for (const line of lines) {
			const createdLines = wrapAnsi(line, max, { hard: true });
			const alignedLines = (0, import_ansi_align.default)(createdLines, { align: textAlignment });
			const alignedLinesArray = alignedLines.split("\n");
			const longestLength = Math.max(...alignedLinesArray.map((s) => stringWidth(s)));
			for (const alignedLine of alignedLinesArray) {
				let paddedLine;
				switch (textAlignment) {
					case "center": {
						paddedLine = PAD.repeat((max - longestLength) / 2) + alignedLine;
						break;
					}
					case "right": {
						paddedLine = PAD.repeat(max - longestLength) + alignedLine;
						break;
					}
					default: {
						paddedLine = alignedLine;
						break;
					}
				}
				newLines.push(paddedLine);
			}
		}
		lines = newLines;
	}
	if (textAlignment === "center" && textWidth < max) lines = lines.map((line) => PAD.repeat((max - textWidth) / 2) + line);
	else if (textAlignment === "right" && textWidth < max) lines = lines.map((line) => PAD.repeat(max - textWidth) + line);
	const paddingLeft = PAD.repeat(padding.left);
	const paddingRight = PAD.repeat(padding.right);
	lines = lines.map((line) => {
		const newLine = paddingLeft + line + paddingRight;
		return newLine + PAD.repeat(width - stringWidth(newLine));
	});
	if (padding.top > 0) lines = [...Array.from({ length: padding.top }).fill(PAD.repeat(width)), ...lines];
	if (padding.bottom > 0) lines = [...lines, ...Array.from({ length: padding.bottom }).fill(PAD.repeat(width))];
	if (height && lines.length > height) lines = lines.slice(0, height);
	else if (height && lines.length < height) lines = [...lines, ...Array.from({ length: height - lines.length }).fill(PAD.repeat(width))];
	return lines.join(NEWLINE);
};
const boxContent = (content, contentWidth, options) => {
	const colorizeBorder = (border) => {
		const newBorder = options.borderColor ? getColorFunction(options.borderColor)(border) : border;
		return options.dimBorder ? source_default.dim(newBorder) : newBorder;
	};
	const colorizeContent = (content$1) => options.backgroundColor ? getBGColorFunction(options.backgroundColor)(content$1) : content$1;
	const chars = getBorderChars(options.borderStyle);
	const columns = terminalColumns();
	let marginLeft = PAD.repeat(options.margin.left);
	if (options.float === "center") {
		const marginWidth = Math.max((columns - contentWidth - getBorderWidth(options.borderStyle)) / 2, 0);
		marginLeft = PAD.repeat(marginWidth);
	} else if (options.float === "right") {
		const marginWidth = Math.max(columns - contentWidth - options.margin.right - getBorderWidth(options.borderStyle), 0);
		marginLeft = PAD.repeat(marginWidth);
	}
	let result = "";
	if (options.margin.top) result += NEWLINE.repeat(options.margin.top);
	if (options.borderStyle !== NONE || options.title) result += colorizeBorder(marginLeft + chars.topLeft + (options.title ? makeTitle(options.title, chars.top.repeat(contentWidth), options.titleAlignment) : chars.top.repeat(contentWidth)) + chars.topRight) + NEWLINE;
	const lines = content.split(NEWLINE);
	result += lines.map((line) => marginLeft + colorizeBorder(chars.left) + colorizeContent(line) + colorizeBorder(chars.right)).join(NEWLINE);
	if (options.borderStyle !== NONE) result += NEWLINE + colorizeBorder(marginLeft + chars.bottomLeft + chars.bottom.repeat(contentWidth) + chars.bottomRight);
	if (options.margin.bottom) result += NEWLINE.repeat(options.margin.bottom);
	return result;
};
const sanitizeOptions = (options) => {
	if (options.fullscreen && process$1?.stdout) {
		let newDimensions = [process$1.stdout.columns, process$1.stdout.rows];
		if (typeof options.fullscreen === "function") newDimensions = options.fullscreen(...newDimensions);
		options.width ||= newDimensions[0];
		options.height ||= newDimensions[1];
	}
	options.width &&= Math.max(1, options.width - getBorderWidth(options.borderStyle));
	options.height &&= Math.max(1, options.height - getBorderWidth(options.borderStyle));
	return options;
};
const formatTitle = (title, borderStyle) => borderStyle === NONE ? title : ` ${title} `;
const determineDimensions = (text, options) => {
	options = sanitizeOptions(options);
	const widthOverride = options.width !== void 0;
	const columns = terminalColumns();
	const borderWidth = getBorderWidth(options.borderStyle);
	const maxWidth = columns - options.margin.left - options.margin.right - borderWidth;
	const widest = widestLine(wrapAnsi(text, columns - borderWidth, {
		hard: true,
		trim: false
	})) + options.padding.left + options.padding.right;
	if (options.title && widthOverride) {
		options.title = options.title.slice(0, Math.max(0, options.width - 2));
		options.title &&= formatTitle(options.title, options.borderStyle);
	} else if (options.title) {
		options.title = options.title.slice(0, Math.max(0, maxWidth - 2));
		if (options.title) {
			options.title = formatTitle(options.title, options.borderStyle);
			if (stringWidth(options.title) > widest) options.width = stringWidth(options.title);
		}
	}
	options.width ||= widest;
	if (!widthOverride) {
		if (options.margin.left && options.margin.right && options.width > maxWidth) {
			const spaceForMargins = columns - options.width - borderWidth;
			const multiplier = spaceForMargins / (options.margin.left + options.margin.right);
			options.margin.left = Math.max(0, Math.floor(options.margin.left * multiplier));
			options.margin.right = Math.max(0, Math.floor(options.margin.right * multiplier));
		}
		options.width = Math.min(options.width, columns - borderWidth - options.margin.left - options.margin.right);
	}
	if (options.width - (options.padding.left + options.padding.right) <= 0) {
		options.padding.left = 0;
		options.padding.right = 0;
	}
	if (options.height && options.height - (options.padding.top + options.padding.bottom) <= 0) {
		options.padding.top = 0;
		options.padding.bottom = 0;
	}
	return options;
};
const isHex = (color) => color.match(/^#(?:[0-f]{3}){1,2}$/i);
const isColorValid = (color) => typeof color === "string" && (source_default[color] ?? isHex(color));
const getColorFunction = (color) => isHex(color) ? source_default.hex(color) : source_default[color];
const getBGColorFunction = (color) => isHex(color) ? source_default.bgHex(color) : source_default[camelCase(["bg", color])];
function boxen(text, options) {
	options = {
		padding: 0,
		borderStyle: "single",
		dimBorder: false,
		textAlignment: "left",
		float: "left",
		titleAlignment: "left",
		...options
	};
	if (options.align) options.textAlignment = options.align;
	if (options.borderColor && !isColorValid(options.borderColor)) throw new Error(`${options.borderColor} is not a valid borderColor`);
	if (options.backgroundColor && !isColorValid(options.backgroundColor)) throw new Error(`${options.backgroundColor} is not a valid backgroundColor`);
	options.padding = getObject(options.padding);
	options.margin = getObject(options.margin);
	options = determineDimensions(text, options);
	text = makeContentText(text, options);
	return boxContent(text, options.width, options);
}

//#endregion
//#region src/lib/init.ts
function ensureGitIgnore() {
	const gitignorePath = join(env.cwd, ".gitignore");
	if (!existsSync(gitignorePath)) {
		writeFileSync(gitignorePath, ".coder\n");
		return;
	}
	const content = readFileSync(gitignorePath, "utf8");
	if (!content.includes(".coder")) appendFileSync(gitignorePath, ".coder\n");
}
async function initConfig() {
	const isTypescript = await mod_default.confirm("Do you want to use TypeScript? (config.config.ts)", { default: false });
	const file = isTypescript ? "config.config.ts" : "config.config.js";
	const coderConfig = join(env.cwd, file);
	if (existsSync(coderConfig)) return;
	if (isTypescript) {
		const pm = await detect({ cwd: env.cwd });
		if (!pm) throw new Error("No package manager detected");
		if (pm?.name === "npm") await mod_default`npm install opencoder@latest`.printCommand().text();
		if (pm?.name === "yarn") await mod_default`yarn add opencoder@latest`.printCommand().text();
		if (pm?.name === "pnpm") await mod_default`pnpm add opencoder@latest`.printCommand().text();
		if (pm?.name === "bun") await mod_default`bun add opencoder@latest`.printCommand().text();
		writeFileSync(coderConfig, dedent_default`
    import { openai, type Config } from "opencoder"

    export const config = {
      // add your custom model simple like this
      // model: openai("gpt-4o")
    } satisfies Config
    `);
	} else writeFileSync(coderConfig, dedent_default`
      export default {
        // add your custom model simple like this
        // model: openai("gpt-4o")
      }
      `);
	console.log(boxen(`Coder config file created at ${source_default.green(join(env.cwd, file))}`, {
		padding: 1,
		borderColor: "green",
		borderStyle: "round"
	}));
}

//#endregion
//#region node_modules/dotenv/package.json
var package_exports = {};
__export(package_exports, {
	browser: () => browser,
	default: () => package_default$1,
	description: () => description,
	devDependencies: () => devDependencies,
	engines: () => engines,
	exports: () => exports,
	funding: () => funding,
	homepage: () => homepage,
	keywords: () => keywords,
	license: () => license,
	main: () => main,
	name: () => name,
	readmeFilename: () => readmeFilename,
	repository: () => repository,
	scripts: () => scripts,
	types: () => types,
	version: () => version$1
});
var name, version$1, description, main, types, exports, scripts, repository, homepage, funding, keywords, readmeFilename, license, devDependencies, engines, browser, package_default$1;
var init_package = __esm({ "node_modules/dotenv/package.json"() {
	name = "dotenv";
	version$1 = "16.5.0";
	description = "Loads environment variables from .env file";
	main = "lib/main.js";
	types = "lib/main.d.ts";
	exports = {
		".": {
			"types": "./lib/main.d.ts",
			"require": "./lib/main.js",
			"default": "./lib/main.js"
		},
		"./config": "./config.js",
		"./config.js": "./config.js",
		"./lib/env-options": "./lib/env-options.js",
		"./lib/env-options.js": "./lib/env-options.js",
		"./lib/cli-options": "./lib/cli-options.js",
		"./lib/cli-options.js": "./lib/cli-options.js",
		"./package.json": "./package.json"
	};
	scripts = {
		"dts-check": "tsc --project tests/types/tsconfig.json",
		"lint": "standard",
		"pretest": "npm run lint && npm run dts-check",
		"test": "tap run --allow-empty-coverage --disable-coverage --timeout=60000",
		"test:coverage": "tap run --show-full-coverage --timeout=60000 --coverage-report=lcov",
		"prerelease": "npm test",
		"release": "standard-version"
	};
	repository = {
		"type": "git",
		"url": "git://github.com/motdotla/dotenv.git"
	};
	homepage = "https://github.com/motdotla/dotenv#readme";
	funding = "https://dotenvx.com";
	keywords = [
		"dotenv",
		"env",
		".env",
		"environment",
		"variables",
		"config",
		"settings"
	];
	readmeFilename = "README.md";
	license = "BSD-2-Clause";
	devDependencies = {
		"@types/node": "^18.11.3",
		"decache": "^4.6.2",
		"sinon": "^14.0.1",
		"standard": "^17.0.0",
		"standard-version": "^9.5.0",
		"tap": "^19.2.0",
		"typescript": "^4.8.4"
	};
	engines = { "node": ">=12" };
	browser = { "fs": false };
	package_default$1 = {
		name,
		version: version$1,
		description,
		main,
		types,
		exports,
		scripts,
		repository,
		homepage,
		funding,
		keywords,
		readmeFilename,
		license,
		devDependencies,
		engines,
		browser
	};
} });

//#endregion
//#region node_modules/dotenv/lib/main.js
var require_main = __commonJS({ "node_modules/dotenv/lib/main.js"(exports, module) {
	const fs$1 = __require("fs");
	const path$1 = __require("path");
	const os = __require("os");
	const crypto = __require("crypto");
	const packageJson = (init_package(), __toCommonJS(package_exports));
	const version = packageJson.version;
	const LINE = /(?:^|^)\s*(?:export\s+)?([\w.-]+)(?:\s*=\s*?|:\s+?)(\s*'(?:\\'|[^'])*'|\s*"(?:\\"|[^"])*"|\s*`(?:\\`|[^`])*`|[^#\r\n]+)?\s*(?:#.*)?(?:$|$)/gm;
	function parse(src) {
		const obj = {};
		let lines = src.toString();
		lines = lines.replace(/\r\n?/gm, "\n");
		let match;
		while ((match = LINE.exec(lines)) != null) {
			const key = match[1];
			let value = match[2] || "";
			value = value.trim();
			const maybeQuote = value[0];
			value = value.replace(/^(['"`])([\s\S]*)\1$/gm, "$2");
			if (maybeQuote === "\"") {
				value = value.replace(/\\n/g, "\n");
				value = value.replace(/\\r/g, "\r");
			}
			obj[key] = value;
		}
		return obj;
	}
	function _parseVault(options) {
		const vaultPath = _vaultPath(options);
		const result = DotenvModule.configDotenv({ path: vaultPath });
		if (!result.parsed) {
			const err = new Error(`MISSING_DATA: Cannot parse ${vaultPath} for an unknown reason`);
			err.code = "MISSING_DATA";
			throw err;
		}
		const keys = _dotenvKey(options).split(",");
		const length = keys.length;
		let decrypted;
		for (let i = 0; i < length; i++) try {
			const key = keys[i].trim();
			const attrs = _instructions(result, key);
			decrypted = DotenvModule.decrypt(attrs.ciphertext, attrs.key);
			break;
		} catch (error) {
			if (i + 1 >= length) throw error;
		}
		return DotenvModule.parse(decrypted);
	}
	function _warn(message) {
		console.log(`[dotenv@${version}][WARN] ${message}`);
	}
	function _debug(message) {
		console.log(`[dotenv@${version}][DEBUG] ${message}`);
	}
	function _dotenvKey(options) {
		if (options && options.DOTENV_KEY && options.DOTENV_KEY.length > 0) return options.DOTENV_KEY;
		if (process.env.DOTENV_KEY && process.env.DOTENV_KEY.length > 0) return process.env.DOTENV_KEY;
		return "";
	}
	function _instructions(result, dotenvKey) {
		let uri;
		try {
			uri = new URL(dotenvKey);
		} catch (error) {
			if (error.code === "ERR_INVALID_URL") {
				const err = new Error("INVALID_DOTENV_KEY: Wrong format. Must be in valid uri format like dotenv://:key_1234@dotenvx.com/vault/.env.vault?environment=development");
				err.code = "INVALID_DOTENV_KEY";
				throw err;
			}
			throw error;
		}
		const key = uri.password;
		if (!key) {
			const err = new Error("INVALID_DOTENV_KEY: Missing key part");
			err.code = "INVALID_DOTENV_KEY";
			throw err;
		}
		const environment = uri.searchParams.get("environment");
		if (!environment) {
			const err = new Error("INVALID_DOTENV_KEY: Missing environment part");
			err.code = "INVALID_DOTENV_KEY";
			throw err;
		}
		const environmentKey = `DOTENV_VAULT_${environment.toUpperCase()}`;
		const ciphertext = result.parsed[environmentKey];
		if (!ciphertext) {
			const err = new Error(`NOT_FOUND_DOTENV_ENVIRONMENT: Cannot locate environment ${environmentKey} in your .env.vault file.`);
			err.code = "NOT_FOUND_DOTENV_ENVIRONMENT";
			throw err;
		}
		return {
			ciphertext,
			key
		};
	}
	function _vaultPath(options) {
		let possibleVaultPath = null;
		if (options && options.path && options.path.length > 0) if (Array.isArray(options.path)) {
			for (const filepath of options.path) if (fs$1.existsSync(filepath)) possibleVaultPath = filepath.endsWith(".vault") ? filepath : `${filepath}.vault`;
		} else possibleVaultPath = options.path.endsWith(".vault") ? options.path : `${options.path}.vault`;
		else possibleVaultPath = path$1.resolve(process.cwd(), ".env.vault");
		if (fs$1.existsSync(possibleVaultPath)) return possibleVaultPath;
		return null;
	}
	function _resolveHome(envPath) {
		return envPath[0] === "~" ? path$1.join(os.homedir(), envPath.slice(1)) : envPath;
	}
	function _configVault(options) {
		const debug = Boolean(options && options.debug);
		if (debug) _debug("Loading env from encrypted .env.vault");
		const parsed = DotenvModule._parseVault(options);
		let processEnv = process.env;
		if (options && options.processEnv != null) processEnv = options.processEnv;
		DotenvModule.populate(processEnv, parsed, options);
		return { parsed };
	}
	function configDotenv(options) {
		const dotenvPath = path$1.resolve(process.cwd(), ".env");
		let encoding = "utf8";
		const debug = Boolean(options && options.debug);
		if (options && options.encoding) encoding = options.encoding;
		else if (debug) _debug("No encoding is specified. UTF-8 is used by default");
		let optionPaths = [dotenvPath];
		if (options && options.path) if (!Array.isArray(options.path)) optionPaths = [_resolveHome(options.path)];
		else {
			optionPaths = [];
			for (const filepath of options.path) optionPaths.push(_resolveHome(filepath));
		}
		let lastError;
		const parsedAll = {};
		for (const path$2 of optionPaths) try {
			const parsed = DotenvModule.parse(fs$1.readFileSync(path$2, { encoding }));
			DotenvModule.populate(parsedAll, parsed, options);
		} catch (e) {
			if (debug) _debug(`Failed to load ${path$2} ${e.message}`);
			lastError = e;
		}
		let processEnv = process.env;
		if (options && options.processEnv != null) processEnv = options.processEnv;
		DotenvModule.populate(processEnv, parsedAll, options);
		if (lastError) return {
			parsed: parsedAll,
			error: lastError
		};
		else return { parsed: parsedAll };
	}
	function config$1(options) {
		if (_dotenvKey(options).length === 0) return DotenvModule.configDotenv(options);
		const vaultPath = _vaultPath(options);
		if (!vaultPath) {
			_warn(`You set DOTENV_KEY but you are missing a .env.vault file at ${vaultPath}. Did you forget to build it?`);
			return DotenvModule.configDotenv(options);
		}
		return DotenvModule._configVault(options);
	}
	function decrypt(encrypted, keyStr) {
		const key = Buffer.from(keyStr.slice(-64), "hex");
		let ciphertext = Buffer.from(encrypted, "base64");
		const nonce = ciphertext.subarray(0, 12);
		const authTag = ciphertext.subarray(-16);
		ciphertext = ciphertext.subarray(12, -16);
		try {
			const aesgcm = crypto.createDecipheriv("aes-256-gcm", key, nonce);
			aesgcm.setAuthTag(authTag);
			return `${aesgcm.update(ciphertext)}${aesgcm.final()}`;
		} catch (error) {
			const isRange = error instanceof RangeError;
			const invalidKeyLength = error.message === "Invalid key length";
			const decryptionFailed = error.message === "Unsupported state or unable to authenticate data";
			if (isRange || invalidKeyLength) {
				const err = new Error("INVALID_DOTENV_KEY: It must be 64 characters long (or more)");
				err.code = "INVALID_DOTENV_KEY";
				throw err;
			} else if (decryptionFailed) {
				const err = new Error("DECRYPTION_FAILED: Please check your DOTENV_KEY");
				err.code = "DECRYPTION_FAILED";
				throw err;
			} else throw error;
		}
	}
	function populate(processEnv, parsed, options = {}) {
		const debug = Boolean(options && options.debug);
		const override = Boolean(options && options.override);
		if (typeof parsed !== "object") {
			const err = new Error("OBJECT_REQUIRED: Please check the processEnv argument being passed to populate");
			err.code = "OBJECT_REQUIRED";
			throw err;
		}
		for (const key of Object.keys(parsed)) if (Object.prototype.hasOwnProperty.call(processEnv, key)) {
			if (override === true) processEnv[key] = parsed[key];
			if (debug) if (override === true) _debug(`"${key}" is already defined and WAS overwritten`);
			else _debug(`"${key}" is already defined and was NOT overwritten`);
		} else processEnv[key] = parsed[key];
	}
	const DotenvModule = {
		configDotenv,
		_configVault,
		_parseVault,
		config: config$1,
		decrypt,
		parse,
		populate
	};
	module.exports.configDotenv = DotenvModule.configDotenv;
	module.exports._configVault = DotenvModule._configVault;
	module.exports._parseVault = DotenvModule._parseVault;
	module.exports.config = DotenvModule.config;
	module.exports.decrypt = DotenvModule.decrypt;
	module.exports.parse = DotenvModule.parse;
	module.exports.populate = DotenvModule.populate;
	module.exports = DotenvModule;
} });
var import_main = __toESM(require_main());

//#endregion
//#region src/main.tsx
import_main.default.config();
ensureGitIgnore();
const [command] = process.argv.slice(2);
if (command === "init") {
	await initConfig();
	process.exit(0);
}
console.log(boxen(`Welcome to OpenCoder@${package_default.version} (Kthulu Foundry Edition)
Model: ${source_default.green(config.model?.modelId || "claude-3-5-sonnet-20241022")}
Working directory: ${source_default.green(env.cwd)}
Kthulu CLI: ${source_default.blue("Integrated")}`, {
	padding: 1,
	borderColor: "cyan",
	borderStyle: "round"
}));
createCoder(config, command);

//#endregion