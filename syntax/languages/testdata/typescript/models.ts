export type Id = string | number | bigint;

interface ApiPage<T extends object> {
	readonly items: T[];
	next?: string | null;
	meta: {
		count: number;
		labels: readonly string[];
	};
}

type MutableDraft<T> = {
	-readonly [K in keyof T]+?: T[K] | undefined;
};

type EventName<T extends string> = `${T}:created` | `${T}:updated`;

const samplePage: ApiPage<{ id: Id; name: string }> = {
	items: [{ id: 42n, name: "alpha" }],
	next: null,
	meta: { count: 1_000, labels: ["new", "trial"] },
};
