class SexyvoiceCli < Formula
  desc "Command-line interface for the SexyVoice speech API"
  homepage "https://sexyvoice.ai"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/gianpaj/sexyvoice-cli/releases/download/v#{version}/sexyvoice-darwin-arm64.tar.gz"
      sha256 "57de10e691003d410eeed99b3a66536a375d4fe1a003502eb6be7b3af8a36cdd"
    else
      url "https://github.com/gianpaj/sexyvoice-cli/releases/download/v#{version}/sexyvoice-darwin-amd64.tar.gz"
      sha256 "451e935e39ead9ca97fad9d5864b9f66e750bbafe59c45af34e874db1705e38a"
    end
  end

  on_linux do
    url "https://github.com/gianpaj/sexyvoice-cli/releases/download/v#{version}/sexyvoice-linux-amd64.tar.gz"
    sha256 "7f03e7c778a0e12e497c55e3980b0ba11e14ff1659d81626fc382f31b25e7570"
  end

  def install
    bin.install "sexyvoice"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/sexyvoice --version")
  end
end
