class Wifiadb < Formula
  desc "Simple scripts that help to ease handy work daily, most of these cli tools was written in Golang"
  homepage "http://hacktons.cn/homebrew-cli"
  url "https://github.com/hacktons/homebrew-cli/releases/download/v0.0.2/wifiadb_0.0.1_macOS_64-bit.tar.gz"
  sha256 "74ed76cacde956d205fef8901a171559c284efdba718ae437584d313ef1330f3"

  version "0.0.1"

  def install
    bin.install "wifiadb"
  end
end