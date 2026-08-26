defmodule Store.Inventory do
  @moduledoc """
  Tracks items and their quantities.
  """

  @default_quantity 0

  defstruct items: %{}, name: :unnamed

  @doc "Add a quantity of an item to the inventory."
  @spec add(t(), atom(), non_neg_integer()) :: t()
  def add(%__MODULE__{items: items} = inventory, item, amount \\ 1)
      when is_atom(item) and amount > 0 do
    updated = Map.update(items, item, amount, &(&1 + amount))
    %{inventory | items: updated}
  end

  def count(%__MODULE__{items: items}, item) do
    Map.get(items, item, @default_quantity)
  end

  def labels(%__MODULE__{items: items}) do
    items
    |> Map.keys()
    |> Enum.map(fn item -> ~s(item:#{item}) end)
    |> Enum.sort()
  end

  # Numbers come in a few shapes.
  def constants do
    [0xFF, 0o17, 0b1010, 1_000_000, 3.14, 1.0e-3, ?A]
  end

  def shout(name) when is_binary(name) do
    "HELLO, " <> String.upcase(name) <> "!"
  end
end
