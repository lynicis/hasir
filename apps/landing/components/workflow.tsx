import { WorkflowStep } from "./workflow-section";

export function WorkflowSection() {
  return (
    <section id="workflow" className="hl-section--surface">
      <div className="container">
        <div className="hl-head-block">
          <h2>From commit to client SDK, automatically.</h2>
        </div>

        <ol className="hl-steps-rail">
          <WorkflowStep
            number="01"
            title="Push your schema"
            description="Commit .proto files and push over Git-SSH, just like any other repository."
          />
          <WorkflowStep
            number="02"
            title="Buf validates on receipt"
            description="Breaking-change detection and lint rules run before anything is accepted."
          />
          <WorkflowStep
            number="03"
            title="SDKs ship themselves"
            description="Worker pools generate and publish client SDKs for TypeScript, JavaScript, and Go."
          />
        </ol>

        <div
          className="hl-pipeline-spec"
          data-slot="workflow"
          data-kind="product"
          data-alt="Hasir server-side buf validation rules and buf.gen.yaml SDK generation pipeline"
        >
          <div className="hl-pipeline-spec__pane">
            <div className="hl-pipeline-spec__head">
              <span className="text-ink font-medium">buf.gen.yaml</span>
              <span className="text-muted">Repository config</span>
            </div>
            <pre className="hl-pipeline-spec__code text-muted">
              <span className="text-ink">version:</span> v2{"\n"}
              <span className="text-ink">plugins:</span>
              {"\n"}
              {"  "}- <span className="text-ink">remote:</span>{" "}
              buf.build/protocolbuffers/go{"\n"}
              {"  "}- <span className="text-ink">remote:</span>{" "}
              buf.build/community/typescript{"\n"}
              <span className="text-accent">
                {"  "}+ hasir: auto-publish on push
              </span>
            </pre>
          </div>

          <div className="hl-pipeline-spec__pane">
            <div className="hl-pipeline-spec__head">
              <span className="text-ink font-medium">
                Server pre-receive gate
              </span>
              <span className="text-accent font-mono tabular-nums">
                0 breaking
              </span>
            </div>
            <pre className="hl-pipeline-spec__code text-muted">
              <span className="text-ink">buf lint</span> STANDARD{"\n"}
              <span className="text-ink">buf breaking</span> --against
              .git#branch=main{"\n"}
              <span className="text-accent">
                pass WIRE_JSON compatibility verified
              </span>
              {"\n"}
              worker-pool: queued go-connectrpc, js-connectrpc
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
