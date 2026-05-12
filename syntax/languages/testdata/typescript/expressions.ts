const flags = 0b1010_0101 & 0xff;
let total = 6.022e23 / 3.0 + 0o755 - 0xdead_beef;

const enabled = (flags & 0b0100) !== 0 && total >= 1e6;
const label = enabled ? `total=${total.toFixed(2)}` : "disabled";

const result = {
	...{ label },
	count: total++,
	path: config?.routes?.[0] ?? "/",
};

total **= 2;
total >>>= 1;
total ||= 10;
total &&= Number.parseFloat("12.5");
