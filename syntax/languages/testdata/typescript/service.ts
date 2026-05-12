abstract class Repository<T extends { id: string }> {
	protected cache = new Map<string, T>();

	constructor(private readonly name: string) {}

	abstract load(id: string): Promise<T | undefined>;

	async findOrCreate(id: string, fallback: () => T): Promise<T> {
		const cached = this.cache.get(id);
		if (cached !== undefined) {
			return cached;
		}

		const loaded = await this.load(id);
		const value = loaded ?? fallback();
		this.cache.set(id, value);
		return value;
	}
}

class MemoryUsers extends Repository<{ id: string; active: boolean }> {
	async load(id: string) {
		return id === "root" ? { id, active: true } : undefined;
	}
}
