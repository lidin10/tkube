class Tkube < Formula
  desc "Enhanced Teleport kubectl wrapper with auto-authentication"
  homepage "https://github.com/lidin10/tkube"
  url "https://github.com/lidin10/tkube/archive/v1.0.0.tar.gz"
  sha256 "YOUR_SHA256_HERE"
  license "MIT"
  head "https://github.com/lidin10/tkube.git", branch: "main"

  depends_on "go" => :build

  def install
    # Build with version information embedded
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags)

    # Generate shell completions
    generate_completions_from_executable(bin/"tkube", "completion")
  end

  test do
    # Test version command
    assert_match version.to_s, shell_output("#{bin}/tkube version")
    
    # Test help command
    assert_match "Enhanced Teleport kubectl wrapper", shell_output("#{bin}/tkube --help")
    
    # Test config path command
    assert_match ".tkube/config.json", shell_output("#{bin}/tkube config path")
    
    # Test status command (should handle missing config gracefully)
    output = shell_output("#{bin}/tkube status", 0)
    assert_match "Configuration file", output
  end
end