defmodule Sigils do
  def examples do
    words = ~w[alpha beta gamma]
    pattern = ~r/\d{4}-\d{2}-\d{2}/
    charlist = ~c(hello)
    doc = ~S"""
    Raw #{not_interpolated} heredoc.
    """

    {words, pattern, charlist, doc}
  end
end
