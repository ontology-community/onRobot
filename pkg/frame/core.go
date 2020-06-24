/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package frame

import (
	"time"

	"github.com/ontio/ontology/common/log"
)

var (
	OntTool   = NewOntologyTool()
	startTime = time.Now().Unix()
)

type Method func() bool
type GcFunc func()

type OntologyTool struct {
	//Map name to method
	methodsMap map[string]Method
	//Map method result
	methodsRes map[string]bool
	//gc func
	gc GcFunc
}

func NewOntologyTool() *OntologyTool {
	return &OntologyTool{
		methodsMap: make(map[string]Method, 0),
		methodsRes: make(map[string]bool, 0),
	}
}

func (this *OntologyTool) RegMethod(name string, method Method) {
	this.methodsMap[name] = method
}

func (this *OntologyTool) RegGCFunc(fn GcFunc) {
	this.gc = fn
}

//Start run
func (this *OntologyTool) Start(methodsList []string) {
	if len(methodsList) > 0 {
		this.runMethodList(methodsList)
		return
	}
	log.Info("No method to run")
	return
}

func (this *OntologyTool) runMethodList(methodsList []string) {
	this.onStart()
	defer this.onFinish(methodsList)

	var rest = func(index int) {
		n := len(methodsList)
		if n > 1 && index < n - 1 {
			time.Sleep(5 * time.Second)
		}
	}

	for i, method := range methodsList {
		this.runMethod(i+1, method)
		rest(i)
	}
}

func (this *OntologyTool) runMethod(index int, methodName string) {
	this.onBeforeMethodStart(index, methodName)
	method := this.getMethodByName(methodName)
	if method != nil {
		ok := method()
		this.onAfterMethodFinish(index, methodName, ok)
		this.methodsRes[methodName] = ok
		this.gc()
	}
}

func (this *OntologyTool) onStart() {
	log.Info("===============================================================")
	log.Info("-------Ontology Tool Start-------")
	log.Info("===============================================================")
	log.Info("")
}

func (this *OntologyTool) onFinish(methodsList []string) {
	failedList := make([]string, 0)
	successList := make([]string, 0)
	for methodName, ok := range this.methodsRes {
		if ok {
			successList = append(successList, methodName)
		} else {
			failedList = append(failedList, methodName)
		}
	}

	skipList := make([]string, 0)
	for _, method := range methodsList {
		_, ok := this.methodsRes[method]
		if !ok {
			skipList = append(skipList, method)
		}
	}

	succCount := len(successList)
	failedCount := len(failedList)
	endTime := time.Now().Unix()

	log.Info("===============================================================")
	log.Infof("Ontology Tool Finish Total:%v Success:%v Failed:%v Skip:%v, SpendTime:%d sec",
		len(methodsList),
		succCount,
		failedCount,
		len(methodsList)-succCount-failedCount,
		endTime-startTime,
	)

	if succCount > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Success list:")
		for i, succ := range successList {
			log.Infof("%d.\t%s", i+1, succ)
		}
	}
	if failedCount > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Fail list:")
		for i, fail := range failedList {
			log.Infof("%d.\t%s", i+1, fail)
		}
	}
	if len(skipList) > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Skip list:")
		for i, skip := range skipList {
			log.Infof("%d.\t%s", i+1, skip)
		}
	}
	log.Info("===============================================================")
}

func (this *OntologyTool) onBeforeMethodStart(index int, methodName string) {
	log.Info("===============================================================")
	log.Infof("%d. Start Method:%s", index, methodName)
	log.Info("---------------------------------------------------------------")
}

func (this *OntologyTool) onAfterMethodFinish(index int, methodName string, res bool) {
	if res {
		log.Infof("Run Method:%s success.", methodName)
	} else {
		log.Infof("Run Method:%s failed.", methodName)
	}
	log.Info("---------------------------------------------------------------")
	log.Info("")
}

func (this *OntologyTool) getMethodByName(name string) Method {
	return this.methodsMap[name]
}
