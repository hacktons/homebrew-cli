class Apkcompare < Formula
  desc "Simple scripts that help to ease handy work daily, most of these cli tools was written in Golang"
  homepage "http://hacktons.cn/homebrew-cli"
  url "https://github.com/hacktons/homebrew-cli/releases/download/v0.0.3/apkcompare_0.0.3_macOS_64-bit.tar.gz"
  sha256 "c81bf9624e8396e146458f519c7b4628fda1dcc39063af0f91e17508bdface6f"

  version "0.0.1"

  def install
    bin.install "apkcompare"
  end
end