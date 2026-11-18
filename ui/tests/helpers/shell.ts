import { execSync } from "child_process";

export const runCommand = (command: string) => {
  console.log("Running command: ", command);
  const result = execSync(command).toString();
  console.log("Result: ", result);
  return result;
};
