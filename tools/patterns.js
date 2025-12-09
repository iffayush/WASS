module.exports = {
    sensitiveFilenames: [
      /\.env$/,
      /docker-compose.*\.ya?ml$/i,
      /aws-credentials/i,
      /credentials/i,
      /id_rsa$/i,
    ],
    secrets: [
      { name: "AWS Access Key ID", re: /\bAKIA[0-9A-Z]{16}\b/g },
      { name: "AWS Secret Access Key", re: /\b[A-Za-z0-9/+=]{40}\b/g },
      { name: "GitHub PAT", re: /\bgh[pousr]_[A-Za-z0-9_]{30,255}\b/g },
      { name: "Likely JWT", re: /\beyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\b/g },
      { name: "Google API Key", re: /\bAIza[0-9A-Za-z\-_]{35}\b/g },
      { name: "Generic API Key", re: /(api[_-]?key|secret|token)\s*[:=]\s*['"]?[A-Za-z0-9\-_\.]{12,128}['"]?/ig },
    ]
  };
  