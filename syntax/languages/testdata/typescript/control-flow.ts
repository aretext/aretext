type Shape =
	| { kind: "circle"; radius: number }
	| { kind: "rect"; width: number; height: number }
	| { kind: "point" };

function area(shape: Shape): number {
	switch (shape.kind) {
	case "circle":
		return Math.PI * shape.radius ** 2;
	case "rect":
		return shape.width * shape.height;
	default:
		return 0;
	}
}

for (const shape of [{ kind: "point" } as Shape]) {
	try {
		console.log(area(shape));
	} catch (err: unknown) {
		throw new Error(`bad shape: ${String(err)}`);
	}
}
