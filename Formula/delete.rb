class Delete < Formula
  desc "Simple scripts that help to ease handy work daily, most of these cli tools was written in Golang"
  homepage "http://hacktons.cn/homebrew-cli"
  url "https://github.com/hacktons/homebrew-cli/releases/download/v0.0.1/deleteBuild_0.0.1_macOS_64-bit.tar.gz"
  sha256 "8de05ac044444d33beca2f27bc274939cdcf2cd1e5f00229c64bf2b1aa02f82a"

  version "0.0.1"

  def install
    bin.install "deleteBuild"
  end

  test do
    system "bin/deleteBuild"
  end
end
