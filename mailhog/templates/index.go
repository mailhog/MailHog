package templates

func Index() string {
	return `
<style>
  .messages {
    height: 30%;
    overflow-y: scroll;
  }
  .preview {
    height: 70%;
    border-top: 1px solid #CCCCCC;
  }
  .preview #headers {
    border-bottom: 1px solid #DDDDDD;
  }
  .selected {
    background: #DADAFA;
  }
  table tbody {
    overflow: scroll;
  }
  table td, table th {
    padding: 2px 4px 2px 4px !important;
  }
  table#headers {
    margin-bottom: 2px;
  }
  #content .nav>li>a {
    padding: 5px 8px;
  }
  #content {
    padding: 0px 2px;
  }
</style>
<div class="modal fade" id="confirm-delete-all">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>
        <h4 class="modal-title">Delete all messages?</h4>
      </div>
      <div class="modal-body">
        <p>Are you sure you want to delete all messages?</p>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
        <button type="button" class="btn btn-danger" ng-click="deleteAllConfirm()">Delete all messages</button>
      </div>
    </div>
  </div>
</div>

<div class="messages">
  <table class="table">
    <tr>
      <th>From</th>
      <th>To</th>
      <th>Subject</th>
      <th>Received</th>
      <th>Actions</th>
    </tr>
    <tbody>
      <tr ng-repeat="message in messages" ng-click="selectMessage(message)" ng-class="{ selected: message.Id == preview.Id }">
        <td>
          {{ message.From.Mailbox }}@{{ message.From.Domain }}
        </td>
        <td>
          <span ng-repeat="to in message.To">
            {{ to.Mailbox }}@{{ to.Domain }}
          </span>
        </td>
        <td>
          {{ message.Content.Headers.Subject.0 }}
        </td>
        <td>
          {{ date(message.Created) }}
        </td>
        <td>
          <button class="btn btn-xs btn-default" title="Delete" ng-click="deleteOne(message)"><span class="glyphicon glyphicon-remove"></span></button>
        </td>
      </tr>
    </tbody>
  </table>
</div>
<div class="preview" ng-if="preview">
  <table class="table" id="headers">
    <tr ng-repeat="(header, value) in preview.Content.Headers">
      <th>
        {{ header }}
      </th>
      <td>
        <div ng-repeat="v in value">{{ v }}</div>
      </td>
    </tr>
  </table>
  <div id="content">
    <ul class="nav nav-tabs">
      <li class="active"><a href="#preview-html" data-toggle="tab">HTML</a></li>
      <li><a href="#preview-plain" data-toggle="tab">Plain text</a></li>
      <li><a href="#preview-source" data-toggle="tab">Source</a></li>
    </ul>
    <div class="tab-content">
      <div class="tab-pane active" id="preview-html" ng-bind-html="preview.previewHTML"></div>
      <div class="tab-pane" id="preview-plain"><pre>{{ getMessagePlain(preview) }}</pre></div>
      <div class="tab-pane" id="preview-source"><pre>{{ getSource(preview) }}</pre></div>
    </div>
  </div>
</div>
`;
}