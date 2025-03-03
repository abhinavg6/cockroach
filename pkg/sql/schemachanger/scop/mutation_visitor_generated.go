// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Code generated by generate_visitor.go. DO NOT EDIT.

package scop

import "context"

// MutationOp is an operation which can be visited by MutationVisitor.
type MutationOp interface {
	Op
	Visit(context.Context, MutationVisitor) error
}

// MutationVisitor is a visitor for MutationOp operations.
type MutationVisitor interface {
	MakeAddedIndexDeleteOnly(context.Context, MakeAddedIndexDeleteOnly) error
	MakeAddedIndexDeleteAndWriteOnly(context.Context, MakeAddedIndexDeleteAndWriteOnly) error
	MakeAddedPrimaryIndexPublic(context.Context, MakeAddedPrimaryIndexPublic) error
	MakeDroppedPrimaryIndexDeleteAndWriteOnly(context.Context, MakeDroppedPrimaryIndexDeleteAndWriteOnly) error
	CreateGcJobForDescriptor(context.Context, CreateGcJobForDescriptor) error
	MarkDescriptorAsDroppedSynthetically(context.Context, MarkDescriptorAsDroppedSynthetically) error
	MarkDescriptorAsDropped(context.Context, MarkDescriptorAsDropped) error
	DrainDescriptorName(context.Context, DrainDescriptorName) error
	UpdateRelationDeps(context.Context, UpdateRelationDeps) error
	RemoveColumnDefaultExpression(context.Context, RemoveColumnDefaultExpression) error
	AddTypeBackRef(context.Context, AddTypeBackRef) error
	RemoveRelationDependedOnBy(context.Context, RemoveRelationDependedOnBy) error
	RemoveTypeBackRef(context.Context, RemoveTypeBackRef) error
	MakeAddedColumnDeleteAndWriteOnly(context.Context, MakeAddedColumnDeleteAndWriteOnly) error
	MakeDroppedNonPrimaryIndexDeleteAndWriteOnly(context.Context, MakeDroppedNonPrimaryIndexDeleteAndWriteOnly) error
	MakeDroppedIndexDeleteOnly(context.Context, MakeDroppedIndexDeleteOnly) error
	MakeIndexAbsent(context.Context, MakeIndexAbsent) error
	MakeAddedColumnDeleteOnly(context.Context, MakeAddedColumnDeleteOnly) error
	MakeColumnPublic(context.Context, MakeColumnPublic) error
	MakeDroppedColumnDeleteAndWriteOnly(context.Context, MakeDroppedColumnDeleteAndWriteOnly) error
	MakeDroppedColumnDeleteOnly(context.Context, MakeDroppedColumnDeleteOnly) error
	MakeColumnAbsent(context.Context, MakeColumnAbsent) error
	AddCheckConstraint(context.Context, AddCheckConstraint) error
	AddColumnFamily(context.Context, AddColumnFamily) error
	DropForeignKeyRef(context.Context, DropForeignKeyRef) error
	RemoveSequenceOwnedBy(context.Context, RemoveSequenceOwnedBy) error
}

// Visit is part of the MutationOp interface.
func (op MakeAddedIndexDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeAddedIndexDeleteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeAddedIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeAddedIndexDeleteAndWriteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeAddedPrimaryIndexPublic) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeAddedPrimaryIndexPublic(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeDroppedPrimaryIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeDroppedPrimaryIndexDeleteAndWriteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op CreateGcJobForDescriptor) Visit(ctx context.Context, v MutationVisitor) error {
	return v.CreateGcJobForDescriptor(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MarkDescriptorAsDroppedSynthetically) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MarkDescriptorAsDroppedSynthetically(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MarkDescriptorAsDropped) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MarkDescriptorAsDropped(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op DrainDescriptorName) Visit(ctx context.Context, v MutationVisitor) error {
	return v.DrainDescriptorName(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op UpdateRelationDeps) Visit(ctx context.Context, v MutationVisitor) error {
	return v.UpdateRelationDeps(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op RemoveColumnDefaultExpression) Visit(ctx context.Context, v MutationVisitor) error {
	return v.RemoveColumnDefaultExpression(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op AddTypeBackRef) Visit(ctx context.Context, v MutationVisitor) error {
	return v.AddTypeBackRef(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op RemoveRelationDependedOnBy) Visit(ctx context.Context, v MutationVisitor) error {
	return v.RemoveRelationDependedOnBy(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op RemoveTypeBackRef) Visit(ctx context.Context, v MutationVisitor) error {
	return v.RemoveTypeBackRef(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeAddedColumnDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeAddedColumnDeleteAndWriteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeDroppedNonPrimaryIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeDroppedNonPrimaryIndexDeleteAndWriteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeDroppedIndexDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeDroppedIndexDeleteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeIndexAbsent) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeIndexAbsent(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeAddedColumnDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeAddedColumnDeleteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeColumnPublic) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeColumnPublic(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeDroppedColumnDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeDroppedColumnDeleteAndWriteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeDroppedColumnDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeDroppedColumnDeleteOnly(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op MakeColumnAbsent) Visit(ctx context.Context, v MutationVisitor) error {
	return v.MakeColumnAbsent(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op AddCheckConstraint) Visit(ctx context.Context, v MutationVisitor) error {
	return v.AddCheckConstraint(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op AddColumnFamily) Visit(ctx context.Context, v MutationVisitor) error {
	return v.AddColumnFamily(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op DropForeignKeyRef) Visit(ctx context.Context, v MutationVisitor) error {
	return v.DropForeignKeyRef(ctx, op)
}

// Visit is part of the MutationOp interface.
func (op RemoveSequenceOwnedBy) Visit(ctx context.Context, v MutationVisitor) error {
	return v.RemoveSequenceOwnedBy(ctx, op)
}
