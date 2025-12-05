# Homebrew formula for domaindetails CLI
# To install: brew install simplebytes-com/tap/domaindetails
#
# This formula is published to: https://github.com/simplebytes-com/homebrew-tap

class Domaindetails < Formula
  desc "Domain RDAP and WHOIS lookup CLI tool"
  homepage "https://domaindetails.com"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/simplebytes-com/domaindetails-cli/releases/download/v#{version}/domaindetails-#{version}-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_ARM64"
    else
      url "https://github.com/simplebytes-com/domaindetails-cli/releases/download/v#{version}/domaindetails-#{version}-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/simplebytes-com/domaindetails-cli/releases/download/v#{version}/domaindetails-#{version}-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"
    else
      url "https://github.com/simplebytes-com/domaindetails-cli/releases/download/v#{version}/domaindetails-#{version}-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"
    end
  end

  def install
    bin.install "domaindetails"
  end

  test do
    system "#{bin}/domaindetails", "--version"
  end
end
