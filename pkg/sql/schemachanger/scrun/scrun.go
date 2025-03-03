// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package scrun

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
)

// RunSchemaChangesInTxn executes in-transaction schema changes for the targeted
// state. These are the immediate changes which take place at DDL statement
// execution time (scop.StatementPhase) or when executing COMMIT
// (scop.PreCommitPhase), rather than the asynchronous changes which are done
// by the schema changer job after the transaction commits.
func RunSchemaChangesInTxn(
	ctx context.Context, deps scexec.Dependencies, state scpb.State,
) (scpb.State, error) {
	if len(state) == 0 {
		return nil, nil
	}
	sc, err := scplan.MakePlan(state, scplan.Params{
		ExecutionPhase: deps.Phase(),
		// TODO(ajwerner): Populate the set of new descriptors
	})
	if err != nil {
		return nil, err
	}
	after := state
	for _, s := range sc.Stages {
		if err := scexec.ExecuteOps(ctx, deps, s.Ops); err != nil {
			return nil, err
		}
		after = s.After
	}
	if len(after) == 0 {
		return nil, nil
	}
	return after, nil
}

// CreateSchemaChangeJob builds and enqueues a schema change job for the target
// state at pre-COMMIT time. This also updates the affected descriptors with the
// id of the created job, effectively locking them to prevent any other schema
// changes concurrent to this job's execution.
func CreateSchemaChangeJob(
	ctx context.Context, deps SchemaChangeJobCreationDependencies, state scpb.State,
) (jobspb.JobID, error) {
	if len(state) == 0 {
		return jobspb.InvalidJobID, nil
	}

	targets := make([]*scpb.Target, len(state))
	states := make([]scpb.Status, len(state))
	// TODO(ajwerner): It may be better in the future to have the builder be
	// responsible for determining this set of descriptors. As of the time of
	// writing, the descriptors to be "locked," descriptors that need schema
	// change jobs, and descriptors with schema change mutations all coincide. But
	// there are future schema changes to be implemented in the new schema changer
	// (e.g., RENAME TABLE) for which this may no longer be true.
	descIDSet := catalog.MakeDescriptorIDSet()
	for i := range state {
		targets[i] = state[i].Target
		states[i] = state[i].Status
		// Depending on the element type either a single descriptor ID
		// will exist or multiple (i.e. foreign keys).
		if id := screl.GetDescID(state[i].Element()); id != descpb.InvalidID {
			descIDSet.Add(id)
		}
	}
	descIDs := descIDSet.Ordered()
	jobID, err := deps.TransactionalJobCreator().CreateJob(ctx, jobs.Record{
		Description:   "Schema change job", // TODO(ajwerner): use const
		Statements:    deps.Statements(),
		Username:      deps.User(),
		DescriptorIDs: descIDs,
		Details:       jobspb.NewSchemaChangeDetails{Targets: targets},
		Progress:      jobspb.NewSchemaChangeProgress{States: states},
		RunningStatus: "",
		NonCancelable: false,
	})
	if err != nil {
		return jobspb.InvalidJobID, err
	}
	// Write the job ID to the affected descriptors.
	if err := scexec.UpdateDescriptorJobIDs(
		ctx,
		deps.Catalog(),
		descIDs,
		jobspb.InvalidJobID,
		jobID,
	); err != nil {
		return jobID, err
	}
	return jobID, nil
}

// RunSchemaChangesInJob contains the business logic for the Resume method of a
// declarative schema change job, with the dependencies abstracted away.
func RunSchemaChangesInJob(
	ctx context.Context,
	deps SchemaChangeJobExecutionDependencies,
	jobID jobspb.JobID,
	jobDescriptorIDs []descpb.ID,
	jobDetails jobspb.NewSchemaChangeDetails,
	jobProgress jobspb.NewSchemaChangeProgress,
) error {
	state := makeState(ctx, deps.ClusterSettings(), jobDetails.Targets, jobProgress.States)
	sc, err := scplan.MakePlan(state, scplan.Params{ExecutionPhase: scop.PostCommitPhase})
	if err != nil {
		return err
	}

	if len(sc.Stages) == 0 {
		// In the case where no stage exists, and therefore there's nothing to
		// execute, we still need to open a transaction to remove all references to
		// this schema change job from the descriptors.
		return deps.WithTxnInJob(ctx, func(ctx context.Context, td SchemaChangeJobTxnDependencies) error {
			c := td.ExecutorDependencies().Catalog()
			return scexec.UpdateDescriptorJobIDs(ctx, c, jobDescriptorIDs, jobID, jobspb.InvalidJobID)
		})
	}

	for i, stage := range sc.Stages {
		isLastStage := i == len(sc.Stages)-1
		// Execute each stage in its own transaction.
		if err := deps.WithTxnInJob(ctx, func(ctx context.Context, td SchemaChangeJobTxnDependencies) error {
			execDeps := td.ExecutorDependencies()
			if err := scexec.ExecuteOps(ctx, execDeps, stage.Ops); err != nil {
				return err
			}
			if err := td.UpdateSchemaChangeJob(ctx, func(md jobs.JobMetadata, ju JobProgressUpdater) error {
				pg := md.Progress.GetNewSchemaChange()
				pg.States = makeStatuses(stage.After)
				ju.UpdateProgress(md.Progress)
				return nil
			}); err != nil {
				return err
			}
			if isLastStage {
				// Remove the reference to this schema change job from all affected
				// descriptors in the transaction executing the last stage.
				return scexec.UpdateDescriptorJobIDs(ctx, execDeps.Catalog(), jobDescriptorIDs, jobID, jobspb.InvalidJobID)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func makeStatuses(next scpb.State) []scpb.Status {
	states := make([]scpb.Status, len(next))
	for i := range next {
		states[i] = next[i].Status
	}
	return states
}

func makeState(
	ctx context.Context, sv *cluster.Settings, protos []*scpb.Target, states []scpb.Status,
) scpb.State {
	if len(protos) != len(states) {
		logcrash.ReportOrPanic(ctx, &sv.SV, "unexpected slice size mismatch %d and %d",
			len(protos), len(states))
	}
	ts := make(scpb.State, len(protos))
	for i := range protos {
		ts[i] = &scpb.Node{
			Target: protos[i],
			Status: states[i],
		}
	}
	return ts
}
